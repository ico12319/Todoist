sap.ui.define(
  [
    "sap/ui/core/mvc/Controller",
    "sap/m/MessageToast",
    "sap/m/MessageBox",
    "sap/m/Dialog",
    "sap/m/Button",
    "sap/m/Label",
    "sap/m/ComboBox",
    "sap/ui/core/Item",
    "sap/m/Bar",
    "sap/m/Title",
    "sap/m/Image",
    "sap/ui/model/json/JSONModel",
    "ui5/walkthrough/services/GraphqlClient",
    "ui5/walkthrough/services/CollaboratorsFetcher",
  ],
  function (
    Controller,
    MessageToast,
    MessageBox,
    Dialog,
    Button,
    Label,
    ComboBox,
    Item,
    Bar,
    Title,
    Image,
    JSONModel,
    GraphqlClient,
    CollaboratorsFetcher
  ) {
    "use strict";

    const COLLABS_PAGE_LIMIT = 3;

    return Controller.extend("ui5.walkthrough.controller.Collaborators", {
      onInit() {
        this.getOwnerComponent()
          .getRouter()
          .getRoute("collaborators")
          .attachPatternMatched(this._onListMatched, this);
      },

      async _onListMatched(oEvt) {
        this._sListId = oEvt.getParameter("arguments").list_id;
        await this._loadCollaboratorsForList(this._sListId);
      },

      // ============== Paging: Next / Prev ===================
      onNextPage: async function () {
        const m = this.getOwnerComponent().getModel("collaborators");
        const pi = m.getProperty("/pageInfo") || {};
        if (!pi.hasNextPage) return;

        await CollaboratorsFetcher.fetchNextPageCollaborators(
          this,
          this._sListId,
          COLLABS_PAGE_LIMIT,
          pi.endCursor
        );
        m.setProperty("/paging", {
          mode: "next",
          limit: COLLABS_PAGE_LIMIT,
          after: pi.endCursor,
        });
      },

      onPrevPage: async function () {
        const m = this.getOwnerComponent().getModel("collaborators");
        const pi = m.getProperty("/pageInfo") || {};
        if (!pi.hasPrevPage) return;

        await CollaboratorsFetcher.fetchPrevPageCollaborators(
          this,
          this._sListId,
          COLLABS_PAGE_LIMIT,
          pi.startCursor
        );
        m.setProperty("/paging", {
          mode: "prev",
          limit: COLLABS_PAGE_LIMIT,
          before: pi.startCursor,
        });
      },

      // ============== Core loaders ==========================
      async _loadCollaboratorsForList(listId) {
        try {
          const payload = JSON.stringify({
            query: `query CollaboratorsByList($id: ID!) {
              result: list(id: $id) {
                collaborators(first: ${COLLABS_PAGE_LIMIT}) {
                  data { id email }
                  pageInfo { startCursor endCursor hasPrevPage hasNextPage }
                  totalCount
                }
              }
            }`,
            variables: { id: listId },
          });

          // GraphqlClient.fetch връща data.result -> тук е самият list
          const list = await GraphqlClient.fetch(payload);
          const node = list?.collaborators || {};
          const aUsers = node.data || [];
          const pageInfo = node.pageInfo || {};
          const totalCount = node.totalCount ?? 0;

          this.getOwnerComponent().setModel(
            new JSONModel({
              users: aUsers,
              pageInfo,
              hasNext: !!pageInfo.hasNextPage,
              hasPrev: !!pageInfo.hasPrevPage,
              totalCount,
              paging: {
                mode: "first",
                limit: COLLABS_PAGE_LIMIT,
                after: null,
                before: null,
              },
            }),
            "collaborators"
          );
        } catch (err) {
          console.error(err);
          MessageToast.show("Unable to load collaborators for this list");
        }
      },

      async _reloadCollaboratorsPage() {
        const m = this.getOwnerComponent().getModel("collaborators");
        if (!m) return this._loadCollaboratorsForList(this._sListId);

        const p = m.getProperty("/paging") || {
          mode: "first",
          limit: COLLABS_PAGE_LIMIT,
        };

        if (p.mode === "first") {
          return this._loadCollaboratorsForList(this._sListId);
        }
        if (p.mode === "next") {
          return CollaboratorsFetcher.fetchNextPageCollaborators(
            this,
            this._sListId,
            p.limit || COLLABS_PAGE_LIMIT,
            p.after
          );
        }
        if (p.mode === "prev") {
          return CollaboratorsFetcher.fetchPrevPageCollaborators(
            this,
            this._sListId,
            p.limit || COLLABS_PAGE_LIMIT,
            p.before
          );
        }
        return this._loadCollaboratorsForList(this._sListId);
      },

      // допълни текущата страница до лимита (като при todos)
      async _refillCollaboratorsPageToLimit() {
        const m = this.getOwnerComponent().getModel("collaborators");
        if (!m) return;

        let items = m.getProperty("/users") || [];
        if (items.length >= COLLABS_PAGE_LIMIT) return;

        let pi = m.getProperty("/pageInfo") || {};
        const need = COLLABS_PAGE_LIMIT - items.length;
        const p = m.getProperty("/paging") || { mode: "first" };

        try {
          if (p.mode === "prev" && pi.hasPrevPage) {
            // допълни от предишни
            const payload = JSON.stringify({
              query: `query FillPrev($id: ID!, $last: Int!, $before: ID) {
                result: list(id: $id) {
                  collaborators(last: $last, before: $before) {
                    data { id email }
                    pageInfo { startCursor endCursor hasPrevPage hasNextPage }
                    totalCount
                  }
                }
              }`,
              variables: {
                id: this._sListId,
                last: need,
                before: pi.startCursor || null,
              },
            });

            const list = await GraphqlClient.fetch(payload);
            const node = list?.collaborators || {};
            const extra = node.data || [];

            items = extra.concat(items);
            m.setProperty("/users", items);

            const newPi = node.pageInfo || {};
            m.setProperty("/pageInfo", {
              startCursor: newPi.startCursor || pi.startCursor,
              endCursor: pi.endCursor,
              hasPrevPage: !!newPi.hasPrevPage,
              hasNextPage: !!pi.hasNextPage,
            });
            m.setProperty("/hasPrev", !!newPi.hasPrevPage);
            m.setProperty("/hasNext", !!pi.hasNextPage);
          } else if (pi.hasNextPage) {
            // допълни от следващи
            const payload = JSON.stringify({
              query: `query FillNext($id: ID!, $first: Int!, $after: ID) {
                result: list(id: $id) {
                  collaborators(first: $first, after: $after) {
                    data { id email }
                    pageInfo { startCursor endCursor hasPrevPage hasNextPage }
                    totalCount
                  }
                }
              }`,
              variables: {
                id: this._sListId,
                first: need,
                after: pi.endCursor || null,
              },
            });

            const list = await GraphqlClient.fetch(payload);
            const node = list?.collaborators || {};
            const extra = node.data || [];

            items = items.concat(extra);
            m.setProperty("/users", items);

            const newPi = node.pageInfo || {};
            m.setProperty("/pageInfo", {
              startCursor: pi.startCursor,
              endCursor: newPi.endCursor || pi.endCursor,
              hasPrevPage: !!pi.hasPrevPage,
              hasNextPage: !!newPi.hasNextPage,
            });
            m.setProperty("/hasPrev", !!pi.hasPrevPage);
            m.setProperty("/hasNext", !!newPi.hasNextPage);
          }
        } catch (e) {
          console.warn("Collaborators top-up failed", e);
        }
      },

      onLogout: function () {
        window.location.replace("http://rest-api:3434/logout");
      },

      async _loadAllUsers() {
        try {
          const payload = JSON.stringify({
            query: `query AllUsers($first: Int) {
              result: users(first: $first) {
                data { id email }
                pageInfo { startCursor endCursor hasPrevPage hasNextPage }
                totalCount
              }
            }`,
            variables: { first: 500 },
          });

          const usersPage = await GraphqlClient.fetch(payload);
          const aUsers = usersPage?.data || [];

          this.getOwnerComponent().setModel(
            new JSONModel({
              users: aUsers,
              pageInfo: usersPage.pageInfo || {},
              totalCount: usersPage.totalCount ?? aUsers.length,
            }),
            "allUsers"
          );
        } catch (e) {
          console.error("Load all users failed", e);
          MessageToast.show("Unable to load users");
          this.getOwnerComponent().setModel(
            new JSONModel({ users: [] }),
            "allUsers"
          );
        }
      },

      async _loadAllCollaboratorsFull(listId) {
        const FIRST = 200;
        let after = null;
        let hasNext = true;
        const all = [];

        try {
          while (hasNext) {
            const payload = JSON.stringify({
              query: `query AllCollabs($id: ID!, $first: Int, $after: ID) {
                result: list(id: $id) {
                  collaborators(first: $first, after: $after) {
                    data { id email }
                    pageInfo { endCursor hasNextPage }
                    totalCount
                  }
                }
              }`,
              variables: { id: listId, first: FIRST, after },
            });

            const list = await GraphqlClient.fetch(payload);
            const node = list?.collaborators || {};
            all.push(...(node.data || []));

            const pi = node.pageInfo || {};
            hasNext = !!pi.hasNextPage;
            after = pi.endCursor || null;
          }

          this.getOwnerComponent().setModel(
            new JSONModel({ users: all }),
            "allCollaborators"
          );
        } catch (e) {
          console.error("Load ALL collaborators failed", e);
          this.getOwnerComponent().setModel(
            new JSONModel({ users: [] }),
            "allCollaborators"
          );
        }
      },

      onAddPress: async function () {
        const comp = this.getOwnerComponent();

        const mAll = comp.getModel("allUsers");
        if (!mAll || (mAll.getProperty("/users") || []).length === 0) {
          await this._loadAllUsers();
        }

        const mAllCollabs = comp.getModel("allCollaborators");
        if (
          !mAllCollabs ||
          (mAllCollabs.getProperty("/users") || []).length === 0
        ) {
          await this._loadAllCollaboratorsFull(this._sListId);
        }

        const allUsers = comp.getModel("allUsers").getProperty("/users") || [];
        const collabUsers =
          comp.getModel("allCollaborators").getProperty("/users") || [];

        const collabIds = new Set(collabUsers.map((u) => u.id).filter(Boolean));
        const collabEmails = new Set(
          collabUsers.map((u) => (u.email || "").toLowerCase())
        );

        const eligible = allUsers.filter((u) => {
          const byId = u.id && collabIds.has(u.id);
          const byEmail = u.email && collabEmails.has(u.email.toLowerCase());
          return !(byId || byEmail);
        });

        comp.setModel(new JSONModel({ users: eligible }), "eligibleUsers");

        if (!this._oAddCollabDlg) {
          this._oAddCollabDlg = this._createAddCollaboratorDialog();
        }
        this._oAddCollabDlg.open();
      },

      _createAddCollaboratorDialog() {
        const oDlgModel = new JSONModel({ selectedEmail: "" });

        const oUserPicker = new ComboBox({
          placeholder: "Select user",
          width: "100%",
          selectedKey: "{dlg>/selectedEmail}",
          autocomplete: true,
          items: {
            path: "eligibleUsers>/users",
            template: new Item({
              key: "{eligibleUsers>email}",
              text: "{eligibleUsers>email}",
            }),
          },
        }).addStyleClass("dlgInput");
        oUserPicker.setModel(
          this.getOwnerComponent().getModel("eligibleUsers"),
          "eligibleUsers"
        );

        const oHeader = new Bar({
          contentLeft: [
            new Image({ src: "images/cute_gopher.png" }).addStyleClass(
              "dlgTitleIcon"
            ),
            new Title({ text: "Add Collaborator" }).addStyleClass(
              "dlgTitleText"
            ),
          ],
        });

        const oDlg = new Dialog({
          customHeader: oHeader,
          title: "Add Collaborator",
          contentWidth: "400px",
          content: [
            new Label({ text: "User", labelFor: oUserPicker }),
            oUserPicker,
          ],
          beginButton: new Button({
            text: "Add",
            type: "Emphasized",
            press: async () => {
              const email = oDlgModel.getProperty("/selectedEmail");
              if (!email) {
                MessageToast.show("Please select a user");
                return;
              }
              try {
                const payload = JSON.stringify({
                  query: `mutation AddCollab($input: CollaboratorInput!) {
                    result: addListCollaborator(input: $input) {
                      success
                      user { id email }
                    }
                  }`,
                  variables: {
                    input: { listId: this._sListId, userEmail: email },
                  },
                });

                const out = await GraphqlClient.fetch(payload);
                if (!out || out.success === false)
                  throw new Error("Add collaborator failed");

                await this._reloadCollaboratorsPage();
                await this._refillCollaboratorsPageToLimit();

                await this._loadAllCollaboratorsFull(this._sListId);

                MessageToast.show(`Collaborator "${email}" added`);
                oDlg.close();
              } catch (err) {
                console.error(err);
                MessageToast.show("Error while adding collaborator");
              }
            },
          }).addStyleClass("dlgPrimaryBtn"),
          endButton: new Button({
            text: "Cancel",
            press: () => oDlg.close(),
          }).addStyleClass("dlgPrimaryBtn"),
          afterClose: () => {
            oDlgModel.setProperty("/selectedEmail", "");
          },
        });

        oDlg.setModel(oDlgModel, "dlg");
        return oDlg;
      },

      onDeletePress(oEvt) {
        const oCtx = oEvt.getSource().getBindingContext("collaborators");
        const oUser = oCtx.getObject();

        MessageBox.confirm(`Remove "${oUser.email}" from this list?`, {
          icon: MessageBox.Icon.WARNING,
          actions: [MessageBox.Action.OK, MessageBox.Action.CANCEL],
          emphasizedAction: MessageBox.Action.OK,
          onClose: async (sAct) => {
            if (sAct !== MessageBox.Action.OK) return;

            try {
              const payload = JSON.stringify({
                query: `mutation ($listId:ID!,$userId:ID!){
                  result: deleteListCollaborator(id:$listId, user_id:$userId) { success }
                }`,
                variables: { listId: this._sListId, userId: oUser.id },
              });

              const res = await GraphqlClient.fetch(payload);
              if (!res?.success) throw new Error("Backend returned false");

              await this._reloadCollaboratorsPage();
              await this._refillCollaboratorsPageToLimit();

              await this._loadAllCollaboratorsFull(this._sListId);

              MessageToast.show(`"${oUser.email}" removed`);
            } catch (err) {
              console.error("Remove collaborator failed", err);
              MessageToast.show("Error while removing collaborator");
            }
          },
        });
      },

      onBackPress() {
        const appRoot = sap.ui.require.toUrl("ui5/walkthrough");
        window.location.replace(appRoot + "/index.html");
      },
    });
  }
);
