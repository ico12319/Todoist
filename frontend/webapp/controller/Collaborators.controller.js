sap.ui.define(
  [
    "sap/ui/core/mvc/Controller",
    "sap/m/MessageToast",
    "sap/m/MessageBox",
    "sap/m/Dialog",
    "sap/m/Input",
    "sap/m/Button",
    "sap/m/Label",
    "ui5/walkthrough/services/GraphqlClient",
    "ui5/walkthrough/services/CollaboratorsFetcher",
  ],
  function (
    Controller,
    MessageToast,
    MessageBox,
    Dialog,
    Input,
    Button,
    Label,
    GraphqlClient,
    CollaboratorsFetcher
  ) {
    "use strict";

    return Controller.extend("ui5.walkthrough.controller.Collaborators", {
      /* ====================================================== */
      /* Lifecycle & routing                                    */
      /* ====================================================== */

      onInit() {
        const oRouter = this.getOwnerComponent().getRouter();
        oRouter
          .getRoute("collaborators")
          .attachPatternMatched(this._onListMatched, this);
      },

      async _onListMatched(oEvt) {
        this._sListId = oEvt.getParameter("arguments").list_id;
        await this._loadCollaboratorsForList(this._sListId);
      },

      onNextPage: async function () {
        const oModel = this.getOwnerComponent().getModel("collaborators");
        const pageInfo = oModel.getProperty("/pageInfo");

        if (pageInfo.hasNextPage) {
          const newPageInfo =
            await CollaboratorsFetcher.fetchNextPageCollaborators(
              this,
              this._sListId,
              3,
              pageInfo.endCursor
            );
        }
      },

      onPrevPage: async function () {
        const oModel = this.getOwnerComponent().getModel("collaborators");
        const pageInfo = oModel.getProperty("/pageInfo");

        if (pageInfo.hasPrevPage) {
          const newPageInfo =
            await CollaboratorsFetcher.fetchPrevPageCollaborators(
              this,
              this._sListId,
              3,
              pageInfo.startCursor
            );
        }
      },

      async _loadCollaboratorsForList(listId) {
        try {
          const payload = JSON.stringify({
            query: `query CollaboratorsByList($id: ID!) {
              result: list(id: $id) {
                collaborators(first: 3) {
                  data { 
                  id
                  email
                }
                  pageInfo{
                    startCursor
                    endCursor
                    hasPrevPage
                    hasNextPage
                  }
                    totalCount
                }
              }
            }`,
            variables: { id: listId },
          });

          const res = await GraphqlClient.fetch(payload);
          const node = res?.collaborators || {};
          const aCollaborators = node.data || [];
          const pageInfo = node.pageInfo || {};
          const totalCount = node.totalCount ?? 0;

          this.getOwnerComponent().setModel(
            new sap.ui.model.json.JSONModel({
              users: aCollaborators,
              pageInfo,
              hasNext: !!pageInfo.hasNextPage,
              hasPrev: !!pageInfo.hasPrevPage,
              totalCount,
            }),
            "collaborators"
          );
        } catch (err) {
          MessageToast.show("Unable to load collaborators for this list");
        }
      },

      /* ====================================================== */
      /* Add collaborator                                       */
      /* ====================================================== */

      onAddPress() {
        if (!this._oAddDlg) {
          this._oAddDlg = this._createAddDialog();
        }
        this._oAddDlg.open();
      },

      _createAddDialog() {
        const oEmail = new Input({ placeholder: "E-mail", width: "100%" });

        return new Dialog({
          title: "Add collaborator",
          contentWidth: "400px",
          content: [new Label({ text: "E-mail", labelFor: oEmail }), oEmail],
          beginButton: new Button({
            text: "Add",
            type: "Emphasized",
            press: async () => {
              const email = oEmail.getValue().trim();
              if (!email) {
                MessageToast.show("Enter e-mail");
                return;
              }

              try {
                const payload = JSON.stringify({
                  query: `mutation ($input:CollaboratorInput!){
                          result:addListCollaborator(input:$input){
                              user{
                               id 
                               email 
                               role
                            }
                            }
                        }`,
                  variables: {
                    input: { list_id: this._sListId, user_email: email },
                  },
                });

                const data = await GraphqlClient.fetch(payload);
                const newUser = data?.result?.user;
                if (!newUser?.id) throw new Error("Add failed");

                const oModel = this.getOwnerComponent().getModel("collabs");
                const aUsers = oModel.getProperty("/users") || [];
                aUsers.push(newUser);
                oModel.setProperty("/users", aUsers);

                MessageToast.show(`"${email}" added`);
                this._oAddDlg.close();
              } catch (e) {
                console.error("Add collab failed", e);
                MessageToast.show(`Error adding "${email}"`);
              }
            },
          }),
          endButton: new Button({
            text: "Cancel",
            press: () => this._oAddDlg.close(),
          }),
          afterClose: () => oEmail.setValue(""),
        });
      },

      /* ====================================================== */
      /* Delete collaborator                                    */
      /* ====================================================== */

      onDeletePress(oEvt) {
        const oCtx = oEvt.getSource().getBindingContext("collabs");
        const oUser = oCtx.getObject();

        MessageBox.confirm(`Remove "${oUser.email}" from this list?`, {
          icon: MessageBox.Icon.WARNING,
          actions: [MessageBox.Action.OK, MessageBox.Action.CANCEL],
          emphasizedAction: MessageBox.Action.OK,
          onClose: async (sAct) => {
            if (sAct !== MessageBox.Action.OK) {
              return;
            }

            try {
              const payload = JSON.stringify({
                query: `mutation ($listId:ID!,$userId:ID!){
                        result:deleteListCollaborator(
                          id:$listId,user_id:$userId){
                          success
                        }
                      }`,
                variables: { listId: this._sListId, userId: oUser.id },
              });

              const res = await GraphqlClient.fetch(payload);
              if (!res?.result?.success) {
                throw new Error("Backend returned false");
              }

              const oModel = oCtx.getModel("collabs");
              const aUsers = oModel.getProperty("/users") || [];
              oModel.setProperty(
                "/users",
                aUsers.filter((u) => u.id !== oUser.id)
              );

              MessageToast.show(`"${oUser.email}" removed`);
            } catch (err) {
              console.error("Remove collaborator failed", err);
              MessageToast.show("Error while removing collaborator");
            }
          },
        });
      },
    });
  }
);
