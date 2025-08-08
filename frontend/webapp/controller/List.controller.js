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
    "sap/m/DatePicker",
    "ui5/walkthrough/services/GraphqlClient",
    "ui5/walkthrough/services/TodoFetcher",
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
    DatePicker,
    GraphqlClient,
    TodoFetcher
  ) {
    "use strict";

    return Controller.extend("ui5.walkthrough.controller.List", {
      onInit() {
        const oRouter = this.getOwnerComponent().getRouter();
        oRouter
          .getRoute("list")
          .attachPatternMatched(this._onListMatched, this);
      },

      async _onListMatched(oEvt) {
        this._sListId = oEvt.getParameter("arguments").list_id;
        await this._loadTodosForList(this._sListId);
      },

      fmtDate: function (s) {
        if (!s) return "";
        var oFmt = sap.ui.core.format.DateFormat.getDateInstance({
          style: "medium",
          UTC: true,
        });
        return oFmt.format(new Date(s));
      },

      onNextPage: async function () {
        const oModel = this.getOwnerComponent().getModel("todos");
        const pageInfo = oModel.getProperty("/pageInfo");
        if (pageInfo.hasNextPage) {
          const newPageInfo = await TodoFetcher.fetchNextPageTodos(
            this,
            this._sListId,
            5,
            pageInfo.endCursor
          );
        }
      },

      onPrevPage: async function () {
        const oModel = this.getOwnerComponent().getModel("todos");
        const pageInfo = oModel.getProperty("/pageInfo");

        if (pageInfo.hasPrevPage) {
          const newPageInfo = await TodoFetcher.fetchPrevPageTodos(
            this,
            this._sListId,
            5,
            pageInfo.startCursor
          );
        }
      },

      async _loadTodosForList(listId) {
        try {
          const payload = JSON.stringify({
            query: `query TodosByList($id: ID!) {
        result: list(id: $id) {
          todos(first: 5) {
            data {
             id 
             name 
             description 
             priority 
             status 
             dueData
          }
            pageInfo { 
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
          const node = res?.todos || {};
          const aTodos = node.data || [];
          const pageInfo = node.pageInfo || {};
          const totalCount = node.totalCount ?? 0;

          this.getOwnerComponent().setModel(
            new sap.ui.model.json.JSONModel({
              todos: aTodos,
              pageInfo,
              hasNext: !!pageInfo.hasNextPage,
              hasPrev: !!pageInfo.hasPrevPage,
              totalCount,
            }),
            "todos"
          );
        } catch (err) {
          console.error("Load todos failed", err);
          MessageToast.show("Unable to load todos for this list");
        }
      },

      _toRFC3339FromDatePicker(dp) {
        const d = dp.getDateValue();
        if (!d) return "";
        return new Date(
          Date.UTC(d.getFullYear(), d.getMonth(), d.getDate())
        ).toISOString();
      },

      onAddPress() {
        if (!this._oAddDlg) {
          this._oAddDlg = this._createAddDialog();
        }
        this._oAddDlg.open();
      },

      _createAddDialog() {
        const oName = new Input({
          placeholder: "Name",
          width: "100%",
        }).addStyleClass("dlgInput");
        const oDesc = new Input({
          placeholder: "Description",
          width: "100%",
        }).addStyleClass("dlgInput");
        const oPriortiy = new Input({
          placeholder: "Priority",
          width: "100%",
        }).addStyleClass("dlgInput");

        const oDue = new DatePicker({
          placeholder: "Due date (optional)",
          width: "100%",
          valueFormat: "yyyy-MM-dd",
          displayFormat: "long",
        }).addStyleClass("dlgInput");

        const oHeader = new Bar({
          contentLeft: [
            new Image({ src: "images/cute_gopher.png" }).addStyleClass(
              "dlgTitleIcon"
            ),
            new Title({ text: "Create New Todo" }).addStyleClass(
              "dlgTitleText"
            ),
          ],
        });

        return new Dialog({
          customHeader: oHeader,
          title: "Create New Todo",
          contentWidth: "400px",
          content: [
            new Label({ text: "Name", labelFor: oName }),
            oName,
            new Label({
              text: "Description",
              labelFor: oDesc,
              class: "sapUiTinyMarginTop",
            }),
            oDesc,
            new Label({ text: "Priority", labelFor: oPriortiy }),
            oPriortiy,
            new Label({ text: "Due Date", labelFor: oDue }),
            oDue,
          ],
          beginButton: new Button({
            text: "Create",
            type: "Emphasized",
            press: async () => {
              const name = oName.getValue().trim();
              const desc = oDesc.getValue().trim();
              const priority = oPriortiy.getValue().trim();

              if (!name || !desc || !priority) {
                MessageToast.show(
                  "Name, description and priority are mandatory fields"
                );
                return;
              }

              try {
                const input = {
                  name,
                  description: desc,
                  listId: this._sListId,
                  priority,
                };
                const dueISO = this._toRFC3339FromDatePicker(oDue);
                if (dueISO) input.dueDate = dueISO;

                const payload = JSON.stringify({
                  query: `mutation CreateTodo($input: CreateTodoInput!) {
                    result: createTodo(input: $input) { 
                              id 
                              name 
                              description
                              dueData
                            }
                  }`,
                  variables: {
                    input,
                  },
                });

                const newTodo = await GraphqlClient.fetch(payload);
                if (!newTodo?.id) throw new Error("Create failed");

                const oModel = this.getOwnerComponent().getModel("todos");
                const aData = oModel.getProperty("/todos") || [];
                aData.push(newTodo);
                oModel.setProperty("/todos", aData);

                MessageToast.show(`Todo "${name}" created`);
                this._oAddDlg.close();
              } catch (err) {
                console.error(err);
                MessageToast.show("Error while creating todo");
              }
            },
          }).addStyleClass("dlgPrimaryBtn"),
          endButton: new Button({
            text: "Cancel",
            press: () => this._oAddDlg.close(),
          }).addStyleClass("dlgPrimaryBtn"),
          afterClose: () => {
            oName.setValue("");
            oDesc.setValue("");
            oPriortiy.setValue("");
          },
        });
      },

      onEditPress(oEvt) {
        this._createEditDialog(oEvt).open();
      },

      _createEditDialog(oEvt) {
        const oCtx = oEvt.getSource().getBindingContext("todos");
        const oTodo = oCtx.getObject();

        const oNewName = new Input({
          placeholder: "Name",
          width: "100%",
          value: oTodo.name,
        });

        const oNewDesc = new Input({
          placeholder: "Description",
          width: "100%",
          value: oTodo.description,
        });

        const oNewPriority = new Input({
          placeholder: "Priority",
          width: "100%",
          value: oTodo.priority,
        });

        const oNewStatus = new Input({
          placeholder: "Status",
          width: "100%",
          value: oTodo.status,
        });

        const oDue = new DatePicker({
          placeholder: "Due date (optional)",
          width: "100%",
          valueFormat: "yyyy-MM-dd",
          displayFormat: "long",
        }).addStyleClass("dlgInput");

        const oHeader = new Bar({
          contentLeft: [
            new Image({ src: "images/cute_gopher.png" }).addStyleClass(
              "dlgTitleIcon"
            ),
            new Title({ text: "Create New Todo" }).addStyleClass(
              "dlgTitleText"
            ),
          ],
        });

        return new Dialog({
          customHeader: oHeader,
          title: "Update Todo",
          contentWidth: "500px",
          verticalScrolling: true,
          content: [
            new Label({ text: "New Name", labelFor: oNewName }),
            oNewName,
            new Label({
              text: "New Description",
              labelFor: oNewDesc,
              class: "sapUiTinyMarginTop",
            }),
            oNewDesc,
            new Label({
              text: "New Priority",
              labelFor: oNewPriority,
              class: "sapUiTinyMarginTop",
            }),
            oNewPriority,
            new Label({
              text: "New status",
              labelFor: oNewStatus,
              class: "sapUiTinyMarginTop",
            }),
            oNewStatus,
            new Label({
              text: "New Due Date",
              labelFor: oDue,
              class: "sapUiTinyMarginTop",
            }),
            oDue,
          ],
          beginButton: new Button({
            text: "Edit",
            type: "Emphasized",
            press: async () => {
              const name = oNewName.getValue().trim();
              const desc = oNewDesc.getValue().trim();
              const priority = oNewPriority.getValue().trim();
              const status = oNewStatus.getValue().trim();

              if (!name || !desc || !priority || !status) {
                MessageToast.show("There can't be empty values!");
                return;
              }

              try {
                const input = {
                  name: name,
                  description: desc,
                  priority: priority,
                  status: status,
                };
                const dueISO = this._toRFC3339FromDatePicker(oDue);

                if (dueISO) input.dueDate = dueISO;

                const payload = JSON.stringify({
                  query: `mutation UpdateTodo($id: ID!, $input: UpdateTodoInput!) {
                    result: updateTodo(id: $id, input: $input) { 
                      id 
                      name 
                      description 
                      status
                      priority
                      dueData
                    }
                  }`,
                  variables: {
                    id: oTodo.id,
                    input: input,
                  },
                });

                const updTodo = await GraphqlClient.fetch(payload);
                if (!updTodo?.id) throw new Error("Update failed");

                const oModel = this.getOwnerComponent().getModel("todos");
                const aData = oModel.getProperty("/todos") || [];
                const idx = aData.findIndex((t) => t.id === updTodo.id);
                if (idx >= 0) aData[idx] = updTodo;
                oModel.setProperty("/todos", aData);

                MessageToast.show("Todo was successfully updated");
                oNewName.getParent().close();
              } catch (err) {
                console.error(err);
                MessageToast.show("Error while updating todo");
              }
            },
          }),
          endButton: new Button({
            text: "Cancel",
            press: (e) => e.getSource().getParent().close(),
          }),
        });
      },

     onInfoPress: async function (oEvent) {
  const oButton = oEvent.getSource();
  const oCtx = oButton.getBindingContext("todos");
  if (!oCtx) { MessageToast.show("No row context"); return; }
  const todoId = oCtx.getProperty("id");

  if (!this._oInfoPopover) {
    this._oInfoPopover = new Popover({
      placement: sap.m.PlacementType.Bottom,
      showHeader: false,
      contentWidth: "280px",
      content: new VBox({
        renderType: "Bare",
        items: [
          new HBox({ items: [ new Label({ text: "Created:", width: "7rem" }), new Text({ text: "{/createdAt}" }) ] }),
          new HBox({ items: [ new Label({ text: "Updated:", width: "7rem" }), new Text({ text: "{/lastUpdated}" }) ] }),
          new HBox({ items: [ new Label({ text: "Assignee:", width: "7rem" }), new Text({ text: "{/assignedTo}" }) ] })
        ]
      })
    });
    this.getView().addDependent(this._oInfoPopover);
  }

  // първо показваме loader данни и отваряме
  const oModel = new sap.ui.model.json.JSONModel({
    createdAt: "Loading…",
    lastUpdated: "Loading…",
    assignedTo: "Loading…"
  });
  this._oInfoPopover.setModel(oModel);
  this._oInfoPopover.openBy(oButton);

  try {
    // !!! Вариант A: snake_case полета от бекенда
    const payload = JSON.stringify({
      query: `query GetTodoInfo($id: ID!) {
        result: todo(id: $id) {
          createdAt
          lastUpdated
          assignedTo { email }
        }
      }`,
      variables: { id: todoId }
    });

    const info = (await GraphqlClient.fetch(payload)) || {};
    const fmt = sap.ui.core.format.DateFormat.getDateTimeInstance({ style: "medium" });

    oModel.setData({
      createdAt: info.createdAt   ? fmt.format(new Date(info.createdAt))   : "—",
      lastUpdated: info.lastUpdated ? fmt.format(new Date(info.lastUpdated)) : "—",
      assignedTo: info.assignedTo?.email || "—"
    });
  } catch (e) {
    console.error("GetTodoInfo failed", e);
    oModel.setData({ createdAt: "—", lastUpdated: "—", assignedTo: "—" });
    MessageToast.show("Unable to load todo info");
  }
},

      onDeletePress(oEvt) {
        const oCtx = oEvt.getSource().getBindingContext("todos");
        const oTodo = oCtx.getObject();

        MessageBox.confirm(`Delete todo "${oTodo.name}"?`, {
          icon: MessageBox.Icon.WARNING,
          actions: [MessageBox.Action.OK, MessageBox.Action.CANCEL],
          emphasizedAction: MessageBox.Action.OK,
          onClose: async (sAct) => {
            if (sAct !== MessageBox.Action.OK) return;

            try {
              const payload = JSON.stringify({
                query: `mutation DeleteTodo($id: ID!) {
                  result: deleteTodo(id: $id) { success id }
                }`,
                variables: { id: oTodo.id },
              });

              const res = await GraphqlClient.fetch(payload);
              if (!res?.success) throw new Error("Backend returned false");

              const m = oCtx.getModel("todos");
              const info = m.getProperty("/pageInfo") || {};
              const items = m.getProperty("/todos") || [];
              const LIMIT = 5;

              if (items.length <= 1 && info.hasPrevPage) {
                await TodoFetcher.fetchPrevPageTodos(
                  this,
                  this._sListId,
                  LIMIT,
                  info.startCursor
                );
              } else {
                await TodoFetcher.fetchNextPageTodos(
                  this,
                  this._sListId,
                  LIMIT,
                  info.startCursor
                );
              }

              MessageToast.show(`Todo "${oTodo.name}" deleted`);
            } catch (err) {
              console.error("Delete failed", err);
              MessageToast.show("Error while deleting todo");
            }
          },
        });
      },
    });
  }
);
