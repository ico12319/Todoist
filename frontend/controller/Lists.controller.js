sap.ui.define(
  [
    "sap/ui/core/mvc/Controller",
    "sap/m/MessageToast",
    "sap/m/MessageBox",
    "sap/m/Dialog",
    "sap/m/Input",
    "sap/m/Button",
    "sap/m/Label",
    "sap/m/Popover",
    "sap/m/VBox",
    "sap/m/HBox",
    "sap/m/Text",
    "sap/m/Link",
    "sap/m/Bar",
    "sap/m/Title",
    "sap/m/Image",
    "ui5/walkthrough/services/GraphqlClient",
    "ui5/walkthrough/services/ListFetcher",
    "sap/ui/model/json/JSONModel",
  ],
  function (
    Controller,
    MessageToast,
    MessageBox,
    Dialog,
    Input,
    Button,
    Label,
    Popover,
    VBox,
    HBox,
    Text,
    Link,
    Bar,
    Title,
    Image,
    GraphqlClient,
    ListFetcher,
    JSONModel
  ) {
    "use strict";

    const PAGE_LIMIT = 5;

    return Controller.extend("ui5.walkthrough.controller.Lists", {
      onInit() {
        const comp = this.getOwnerComponent();
        if (!comp.getModel("listFilters")) {
          comp.setModel(new JSONModel({ name: "" }), "listFilters");
        }

        this._doLiveSearch = this._debounce(async (term) => {
          await this._applySearch(term);
        }, 350);

        this._fetchListsPage({ first: PAGE_LIMIT }).catch(() =>
          MessageToast.show("Unable to load lists")
        );
      },

      onSearchLiveChange(oEvent) {
        const term =
          oEvent.getParameter("newValue") ?? oEvent.getSource().getValue();
        this._doLiveSearch(term);
      },

      onSearch(oEvent) {
        const term =
          oEvent.getParameter("query") ?? oEvent.getSource().getValue();
        this._applySearch(term);
      },

      async _applySearch(term) {
        const comp = this.getOwnerComponent();
        const trimmed = (term || "").trim();
        comp.getModel("listFilters").setProperty("/name", trimmed);

        await this._fetchListsPage({ first: PAGE_LIMIT });
        const m = comp.getModel("lists");
        if (m) {
          m.setProperty("/paging", {
            mode: "first",
            limit: PAGE_LIMIT,
            after: null,
            before: null,
          });
        }
      },

      _getListCriteria() {
        const m = this.getOwnerComponent().getModel("listFilters");
        const name = (m && m.getProperty("/name")) || "";
        const t = (name || "").trim();
        return t ? { name: t } : null;
      },

      async _fetchListsPage({ first, after, last, before }) {
        const criteria = this._getListCriteria();

        const payload = JSON.stringify({
          query: `query Lists($first: Int, $after: ID, $last: Int, $before: ID, $criteria: ListFilterInput) {
            result: lists(first: $first, after: $after, last: $last, before: $before, criteria: $criteria) {
              data { id name description }
              pageInfo { startCursor endCursor hasPrevPage hasNextPage }
              totalCount
            }
          }`,
          variables: {
            first: first ?? null,
            after: after ?? null,
            last: last ?? null,
            before: before ?? null,
            criteria,
          },
        });

        const page = await GraphqlClient.fetch(payload);
        const aLists = page?.data || [];
        const pageInfo = page?.pageInfo || {};
        const totalCount = page?.totalCount ?? 0;

        const comp = this.getOwnerComponent();
        const prevPaging = comp.getModel("lists")?.getProperty("/paging");

        comp.setModel(
          new JSONModel({
            lists: aLists,
            pageInfo,
            hasNext: !!pageInfo.hasNextPage,
            hasPrev: !!pageInfo.hasPrevPage,
            totalCount,
            paging: prevPaging || {
              mode: "first",
              limit: PAGE_LIMIT,
              after: null,
              before: null,
            },
          }),
          "lists"
        );
      },

      onNextPage: async function () {
        const m = this.getOwnerComponent().getModel("lists");
        const pi = m.getProperty("/pageInfo");
        if (!pi?.hasNextPage) return;

        await this._fetchListsPage({ first: PAGE_LIMIT, after: pi.endCursor });
        m.setProperty("/paging", {
          mode: "next",
          limit: PAGE_LIMIT,
          after: pi.endCursor,
        });
      },

      onPrevPage: async function () {
        const m = this.getOwnerComponent().getModel("lists");
        const pi = m.getProperty("/pageInfo");
        if (!pi?.hasPrevPage) return;

        await this._fetchListsPage({
          last: PAGE_LIMIT,
          before: pi.startCursor,
        });
        m.setProperty("/paging", {
          mode: "prev",
          limit: PAGE_LIMIT,
          before: pi.startCursor,
        });
      },

      onAddPress: function () {
        if (!this._oAddDlg) this._oAddDlg = this._createAddDialog();
        this._oAddDlg.open();
      },

      onCollaboratorsPress: function (oEvent) {
        const oCtx = oEvent.getSource().getBindingContext("lists");
        const oList = oCtx.getObject();
        this.getOwnerComponent()
          .getRouter()
          .navTo("collaborators", { list_id: oList.id });
      },

      _createAddDialog: function () {
        const oName = new Input({
          placeholder: "Name",
          width: "100%",
        }).addStyleClass("dlgInput");
        const oDesc = new Input({
          placeholder: "Description",
          width: "100%",
        }).addStyleClass("dlgInput");

        const oHeader = new Bar({
          contentLeft: [
            new Image({ src: "images/cute_gopher.png" }).addStyleClass(
              "dlgTitleIcon"
            ),
            new Title({ text: "Create New List" }).addStyleClass(
              "dlgTitleText"
            ),
          ],
        });

        const oDlg = new Dialog({
          customHeader: oHeader,
          contentWidth: "420px",
          draggable: true,
          resizable: true,
          stretchOnPhone: true,
          content: [
            new Label({ text: "Name", labelFor: oName }).addStyleClass(
              "dlgLabel"
            ),
            oName,
            new Label({
              text: "Description",
              labelFor: oDesc,
              class: "sapUiTinyMarginTop",
            }).addStyleClass("dlgLabel"),
            oDesc,
          ],
          beginButton: new Button({
            text: "Create",
            type: "Emphasized",
            press: async () => {
              const name = oName.getValue().trim();
              const desc = oDesc.getValue().trim();
              if (!name || !desc) {
                MessageToast.show("Both fields are required");
                return;
              }
              try {
                const payload = JSON.stringify({
                  query: `mutation CreateList($input: CreateListInput!) {
                    result: createList(input: $input) { id name description }
                  }`,
                  variables: { input: { name, description: desc } },
                });

                const newList = await GraphqlClient.fetch(payload);
                if (!newList?.id) throw new Error("Create failed");

                await this._fetchListsPage({ first: PAGE_LIMIT });
                MessageToast.show(`List "${name}" created`);
                this._oAddDlg.close();
              } catch (err) {
                console.error(err);
                MessageToast.show("Error while creating list");
              }
            },
          }).addStyleClass("dlgPrimaryBtn"),
          endButton: new Button({
            text: "Cancel",
            press: () => this._oAddDlg.close(),
          }).addStyleClass("dlgCancelBtn"),
          afterClose: () => {
            oName.setValue("");
            oDesc.setValue("");
          },
        }).addStyleClass("neoDialog");

        this.getView().addDependent(oDlg);
        return oDlg;
      },

      onEditPress: function (oEvent) {
        this._createEditDialog(oEvent).open();
      },

      _createEditDialog: function (oEvent) {
        const oCtx = oEvent.getSource().getBindingContext("lists");
        const oList = oCtx.getObject();

        const oNewName = new Input({
          placeholder: "Name",
          width: "100%",
          value: oList.name || "",
        }).addStyleClass("dlgInput");
        const oNewDesc = new Input({
          placeholder: "Description",
          width: "100%",
          value: oList.description || "",
        }).addStyleClass("dlgInput");

        const oDlg = new Dialog({
          customHeader: new Bar({
            contentLeft: [
              new Image({ src: "images/waving-gopher.png" }).addStyleClass(
                "dlgTitleIcon"
              ),
              new Title({ text: "Update List" }).addStyleClass("dlgTitleText"),
            ],
          }),
          contentWidth: "500px",
          draggable: true,
          resizable: true,
          stretchOnPhone: true,
          content: [
            new Label({ text: "New Name", labelFor: oNewName }).addStyleClass(
              "dlgLabel"
            ),
            oNewName,
            new Label({
              text: "New Description",
              labelFor: oNewDesc,
              class: "sapUiTinyMarginTop",
            }).addStyleClass("dlgLabel"),
            oNewDesc,
          ],
          beginButton: new Button({
            text: "Edit",
            type: "Emphasized",
            press: async () => {
              const name = oNewName.getValue().trim();
              const desc = oNewDesc.getValue().trim();
              if (!name && !desc) {
                MessageToast.show("You can't use empty values!");
                return;
              }
              try {
                const payload = JSON.stringify({
                  query: `mutation UpdateList($id: ID!, $input: UpdateListInput!) {
                    result: updateList(id: $id, input: $input) { id name description }
                  }`,
                  variables: {
                    id: oList.id,
                    input: { name, description: desc },
                  },
                });

                const updatedList = await GraphqlClient.fetch(payload);
                if (!updatedList?.id) throw new Error("Update failed");

                // Рефреш със същия search критерий
                await this._fetchListsPage({ first: PAGE_LIMIT });
                MessageToast.show("List was successfully updated");
                oNewName.getParent().close();
              } catch (err) {
                console.error(err);
                MessageToast.show("Error while trying to update list");
              }
            },
          }).addStyleClass("dlgPrimaryBtn"),
          endButton: new Button({
            text: "Cancel",
            press: (e) => e.getSource().getParent().close(),
          }).addStyleClass("dlgCancelBtn"),
        }).addStyleClass("neoDialog");

        this.getView().addDependent(oDlg);
        return oDlg;
      },

      onInfoPress: async function (oEvent) {
        const oButton = oEvent.getSource();
        const oCtx = oButton.getBindingContext("lists");
        const oList = oCtx ? oCtx.getObject() : {};

        if (!this._oInfoPopover) {
          this._oInfoPopover = new Popover({
            placement: sap.m.PlacementType.Bottom,
            showHeader: false,
            contentWidth: "260px",
            content: new VBox({
              renderType: "Bare",
              items: [
                new HBox({
                  items: [
                    new Label({ text: "Created:", width: "6rem" }),
                    new Text({ text: "{/createdAt}" }),
                  ],
                }),
                new HBox({
                  items: [
                    new Label({ text: "Updated:", width: "6rem" }),
                    new Text({ text: "{/lastUpdated}" }),
                  ],
                }),
                new HBox({
                  items: [
                    new Label({ text: "By:", width: "6rem" }),
                    new Text({ text: "{/createdBy}" }),
                  ],
                }),
                new Button({
                  text: "More info",
                  type: "Transparent",
                  press: this.onMoreInfo.bind(this),
                }),
              ],
            }),
          });
          this.getView().addDependent(this._oInfoPopover);
        }
        this._oInfoPopover.setBindingContext(oCtx, "lists");

        this._oInfoPopover.setModel(
          new JSONModel({
            createdAt: "Loading…",
            lastUpdated: "Loading…",
            createdBy: "Loading…",
          })
        );
        this._oInfoPopover.openBy(oButton);

        try {
          const payload = JSON.stringify({
            query: `query GetListInfo($id: ID!) {
              result: list(id: $id) { created_at last_updated owner { email } }
            }`,
            variables: { id: oList.id },
          });

          const info = (await GraphqlClient.fetch(payload)) || {};
          const fmt = sap.ui.core.format.DateFormat.getDateTimeInstance({
            style: "medium",
          });

          this._oInfoPopover.getModel().setData({
            createdAt: info.created_at
              ? fmt.format(new Date(info.created_at))
              : "—",
            lastUpdated: info.last_updated
              ? fmt.format(new Date(info.last_updated))
              : "—",
            createdBy: info.owner?.email || "—",
          });
        } catch (err) {
          console.error("GetListInfo failed", err);
          this._oInfoPopover
            .getModel()
            .setData({ createdAt: "—", lastUpdated: "—", createdBy: "Error" });
          MessageToast.show("Unable to load list info");
        }
      },

      onMoreInfo: function (oEvent) {
        const oCtx = oEvent.getSource().getBindingContext("lists");
        const sId = oCtx && oCtx.getProperty("id");
        if (!sId) {
          MessageToast.show("Cannot navigate: missing list ID");
          return;
        }
        this._oInfoPopover.close();
        this.getOwnerComponent().getRouter().navTo("list", { list_id: sId });
      },

      onDeletePress: function (oEvent) {
        const oCtx = oEvent.getSource().getBindingContext("lists");
        const oList = oCtx.getObject();

        MessageBox.confirm(`Delete list "${oList.name}"?`, {
          icon: MessageBox.Icon.WARNING,
          actions: [MessageBox.Action.OK, MessageBox.Action.CANCEL],
          emphasizedAction: MessageBox.Action.OK,
          onClose: async (sAct) => {
            if (sAct !== MessageBox.Action.OK) return;

            try {
              const payload = JSON.stringify({
                query: `mutation DeleteList($id: ID!) {
                  result: deleteList(id: $id) { success id }
                }`,
                variables: { id: oList.id },
              });

              const res = await GraphqlClient.fetch(payload);
              if (!res?.success) throw new Error("Backend returned false");

              await this._fetchListsPage({ first: PAGE_LIMIT });
              MessageToast.show(`List "${oList.name}" deleted`);
            } catch (err) {
              console.error("Delete failed", err);
              MessageToast.show("Error while deleting list");
            }
          },
        });
      },

      _debounce(fn, delay) {
        let t;
        return (...args) => {
          clearTimeout(t);
          t = setTimeout(() => fn.apply(this, args), delay);
        };
      },
    });
  }
);
