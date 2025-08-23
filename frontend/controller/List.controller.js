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
    "sap/m/ComboBox",
    "sap/ui/core/Item",
    "ui5/walkthrough/services/GraphqlClient",
    "ui5/walkthrough/services/TodoFetcher",
    "ui5/walkthrough/services/CollaboratorsFetcher",
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
    DatePicker,
    ComboBox,
    Item,
    GraphqlClient,
    TodoFetcher,
    CollaboratorsFetcher,
    JSONModel
  ) {
    "use strict";

    const PAGE_LIMIT = 5;

    return Controller.extend("ui5.walkthrough.controller.List", {
      onInit() {
        this.getOwnerComponent()
          .getRouter()
          .getRoute("list")
          .attachPatternMatched(this._onListMatched, this);
      },

      async _onListMatched(oEvt) {
        this._sListId = oEvt.getParameter("arguments").list_id;

        try {
          await CollaboratorsFetcher.fetchNextPageCollaborators(
            this,
            this._sListId,
            50,
            null
          );
          await this._ensureOwnerInCollaborators(this._sListId);
        } catch (e) {
          console.warn("Collaborators prefetch failed:", e);
        }

        await this._loadTodosForList(this._sListId);
      },

      async _ensureOwnerInCollaborators(listId) {
        try {
          const payload = JSON.stringify({
            query: `query ListOwner($id: ID!) {
              result: list(id: $id) { owner { id email } }
            }`,
            variables: { id: listId },
          });
          const list = await GraphqlClient.fetch(payload);
          const owner = list?.owner;
          if (!owner) return;

          const oModel = this.getOwnerComponent().getModel("collaborators");
          if (!oModel) return;

          const users = oModel.getProperty("/users") || [];
          const exists = users.some((u) => u.id === owner.id);
          if (!exists) {
            oModel.setProperty("/users", [owner, ...users]);
          }
        } catch (e) {
          console.warn("Failed to add owner to collaborators:", e);
        }
      },

      fmtDate(s) {
        if (!s) return "";
        const oFmt = sap.ui.core.format.DateFormat.getDateInstance({
          style: "medium",
          UTC: true,
        });
        return oFmt.format(new Date(s));
      },

      _initFiltersModel() {
        const comp = this.getOwnerComponent();
        if (!comp.getModel("todoFilters")) {
          comp.setModel(
            new JSONModel({
              status: null,
              priority: null,
              type: null,
              name: "",
            }),
            "todoFilters"
          );
        }
      },

      _getActiveFilter() {
        const m = this.getOwnerComponent().getModel("todoFilters");
        if (!m) return null;
        const f = m.getData() || {};
        const name = (f.name || "").trim();

        if (!f.status && !f.priority && !f.type && !name) return null;

        const filter = {
          status: f.status || null,
          priority: f.priority || null,
          type: f.type || null,
        };
        if (name) filter.name = name;
        return filter;
      },

      async _fetchTodosPage({ first, after, last, before }) {
        const filter = this._getActiveFilter();

        const payload = JSON.stringify({
          query: `query Todos($id: ID!, $first: Int, $after: ID, $last: Int, $before: ID, $filter: TodosFilterInput) {
            result: list(id: $id) {
              todos(first: $first, after: $after, last: $last, before: $before, filter: $filter) {
                data {
                  id name description priority status dueDate
                  assignedTo { id email }
                }
                pageInfo { startCursor endCursor hasPrevPage hasNextPage }
                totalCount
              }
            }
          }`,
          variables: {
            id: this._sListId,
            first: first ?? null,
            after: after ?? null,
            last: last ?? null,
            before: before ?? null,
            filter,
          },
        });

        const list = await GraphqlClient.fetch(payload);
        const node = list?.todos || {};
        const aTodos = node.data || [];
        const pageInfo = node.pageInfo || {};
        const totalCount = node.totalCount ?? 0;

        this.getOwnerComponent().setModel(
          new JSONModel({
            todos: aTodos,
            pageInfo,
            hasNext: !!pageInfo.hasNextPage,
            hasPrev: !!pageInfo.hasPrevPage,
            totalCount,
            paging: this.getOwnerComponent()
              .getModel("todos")
              ?.getProperty("/paging") || {
              mode: "first",
              limit: PAGE_LIMIT,
              after: null,
              before: null,
            },
          }),
          "todos"
        );
        return { aTodos, pageInfo, totalCount };
      },

      async _fetchTodosChunk({ first, after, last, before }) {
        const filter = this._getActiveFilter();
        const payload = JSON.stringify({
          query: `query TodosChunk($id: ID!, $first: Int, $after: ID, $last: Int, $before: ID, $filter: TodosFilterInput) {
            result: list(id: $id) {
              todos(first: $first, after: $after, last: $last, before: $before, filter: $filter) {
                data {
                  id name description priority status dueDate
                  assignedTo { id email }
                }
                pageInfo { startCursor endCursor hasPrevPage hasNextPage }
                totalCount
              }
            }
          }`,
          variables: { id: this._sListId, first, after, last, before, filter },
        });
        const list = await GraphqlClient.fetch(payload);
        const node = list?.todos || {};
        return { data: node.data || [], pageInfo: node.pageInfo || {} };
      },

      onFilterPress(oEvent) {
        this._initFiltersModel();

        if (!this._oFilterPopover) {
          const base = this.getOwnerComponent()
            .getModel("todoFilters")
            .getData();
          const fModel = new JSONModel({
            status: base.status || "",
            priority: base.priority || "",
            overdue: base.type || "",
          });

          const statusCB = new ComboBox({
            width: "100%",
            selectedKey: "{f>/status}",
            placeholder: "Status",
            items: [
              new Item({ key: "", text: "Any status" }),
              new Item({ key: "OPEN", text: "OPEN" }),
              new Item({ key: "IN_PROGRESS", text: "IN_PROGRESS" }),
              new Item({ key: "DONE", text: "DONE" }),
            ],
          }).addStyleClass("dlgInput");

          const priorityCB = new ComboBox({
            width: "100%",
            selectedKey: "{f>/priority}",
            placeholder: "Priority",
            items: [
              new Item({ key: "", text: "Any priority" }),
              new Item({ key: "VERY_LOW", text: "VERY_LOW" }),
              new Item({ key: "LOW", text: "LOW" }),
              new Item({ key: "MEDIUM", text: "MEDIUM" }),
              new Item({ key: "HIGH", text: "HIGH" }),
              new Item({ key: "VERY_HIGH", text: "VERY_HIGH" }),
            ],
          }).addStyleClass("dlgInput");

          const overdueCB = new ComboBox({
            width: "100%",
            selectedKey: "{f>/overdue}",
            placeholder: "Overdue",
            items: [
              new Item({ key: "", text: "All (active & expired)" }),
              new Item({ key: "ACTIVE", text: "Active (not overdue)" }),
              new Item({ key: "EXPIRED", text: "Overdue" }),
            ],
          }).addStyleClass("dlgInput");

          const content = new VBox({
            items: [
              new Label({ text: "Status" }),
              statusCB,
              new Label({ text: "Priority", class: "sapUiTinyMarginTop" }),
              priorityCB,
              new Label({ text: "Overdue", class: "sapUiTinyMarginTop" }),
              overdueCB,
            ],
          }).addStyleClass("sapUiSmallMargin");

          this._oFilterPopover = new Popover({
            showHeader: false,
            placement: sap.m.PlacementType.Bottom,
            contentWidth: "260px",
            content: [content],
            footer: new Bar({
              contentRight: [
                new Button({
                  text: "Clear",
                  type: "Transparent",
                  press: async () => {
                    const comp = this.getOwnerComponent();
                    comp.getModel("todoFilters").setData({
                      status: null,
                      priority: null,
                      type: null,
                      name: "",
                    });
                    await this._fetchTodosPage({ first: PAGE_LIMIT });
                    const m = comp.getModel("todos");
                    m.setProperty("/paging", {
                      mode: "first",
                      limit: PAGE_LIMIT,
                      after: null,
                      before: null,
                    });
                    this._oFilterPopover.close();
                  },
                }),
                new Button({
                  text: "Apply",
                  type: "Emphasized",
                  press: async () => {
                    const sel = fModel.getData();
                    const comp = this.getOwnerComponent();
                    comp.getModel("todoFilters").setData({
                      status: sel.status || null,
                      priority: sel.priority || null,
                      type: sel.overdue || null,
                      name:
                        comp.getModel("todoFilters").getProperty("/name") || "",
                    });

                    await this._fetchTodosPage({ first: PAGE_LIMIT });
                    const m = comp.getModel("todos");
                    m.setProperty("/paging", {
                      mode: "first",
                      limit: PAGE_LIMIT,
                      after: null,
                      before: null,
                    });
                    this._oFilterPopover.close();
                  },
                }),
              ],
            }),
          });

          this._oFilterPopover.setModel(fModel, "f");
          this.getView().addDependent(this._oFilterPopover);
        }

        this._oFilterPopover.openBy(oEvent.getSource());
      },

      onTodoSearchLiveChange(oEvent) {
        const val = (oEvent.getParameter("newValue") || "").trimStart();
        const m = this.getOwnerComponent().getModel("todoFilters");
        if (m) m.setProperty("/name", val);

        clearTimeout(this._todoSearchTimer);
        this._todoSearchTimer = setTimeout(() => this._applyTodoSearch(), 300);
      },

      onTodoSearch() {
        clearTimeout(this._todoSearchTimer);
        this._applyTodoSearch();
      },

      async _applyTodoSearch() {
        await this._fetchTodosPage({ first: PAGE_LIMIT });
        const m = this.getOwnerComponent().getModel("todos");
        m.setProperty("/paging", {
          mode: "first",
          limit: PAGE_LIMIT,
          after: null,
          before: null,
        });
      },

      onNextPage: async function () {
        const m = this.getOwnerComponent().getModel("todos");
        const pi = m.getProperty("/pageInfo") || {};
        if (!pi.hasNextPage) return;
        await this._fetchTodosPage({ first: PAGE_LIMIT, after: pi.endCursor });
        m.setProperty("/paging", {
          mode: "next",
          limit: PAGE_LIMIT,
          after: pi.endCursor,
        });
      },

      onPrevPage: async function () {
        const m = this.getOwnerComponent().getModel("todos");
        const pi = m.getProperty("/pageInfo") || {};
        if (!pi.hasPrevPage) return;
        await this._fetchTodosPage({
          last: PAGE_LIMIT,
          before: pi.startCursor,
        });
        m.setProperty("/paging", {
          mode: "prev",
          limit: PAGE_LIMIT,
          before: pi.startCursor,
        });
      },

      async _reloadCurrentPage() {
        const m = this.getOwnerComponent().getModel("todos");
        if (!m) return this._loadTodosForList(this._sListId);
        const p = m.getProperty("/paging") || {
          mode: "first",
          limit: PAGE_LIMIT,
        };
        if (p.mode === "first")
          return this._fetchTodosPage({ first: PAGE_LIMIT });
        if (p.mode === "next")
          return this._fetchTodosPage({
            first: p.limit || PAGE_LIMIT,
            after: p.after,
          });
        if (p.mode === "prev")
          return this._fetchTodosPage({
            last: p.limit || PAGE_LIMIT,
            before: p.before,
          });
        return this._fetchTodosPage({ first: PAGE_LIMIT });
      },

      async _refillPageToLimit() {
        const m = this.getOwnerComponent().getModel("todos");
        if (!m) return;

        let items = m.getProperty("/todos") || [];
        if (items.length >= PAGE_LIMIT) return;

        let pi = m.getProperty("/pageInfo") || {};
        const need = PAGE_LIMIT - items.length;
        const p = m.getProperty("/paging") || { mode: "first" };

        try {
          if (p.mode === "prev" && pi.hasPrevPage) {
            const { data: extra, pageInfo: newPi } =
              await this._fetchTodosChunk({
                last: need,
                before: pi.startCursor || null,
              });

            items = extra.concat(items);
            m.setProperty("/todos", items);
            m.setProperty("/pageInfo", {
              startCursor: newPi.startCursor || pi.startCursor,
              endCursor: pi.endCursor,
              hasPrevPage: !!newPi.hasPrevPage,
              hasNextPage: !!pi.hasNextPage,
            });
            m.setProperty("/hasPrev", !!newPi.hasPrevPage);
            m.setProperty("/hasNext", !!pi.hasNextPage);
          } else if (pi.hasNextPage) {
            const { data: extra, pageInfo: newPi } =
              await this._fetchTodosChunk({
                first: need,
                after: pi.endCursor || null,
              });

            items = items.concat(extra);
            m.setProperty("/todos", items);
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
          console.warn("Top-up failed", e);
        }
      },

      async _loadTodosForList() {
        this._initFiltersModel();
        await this._fetchTodosPage({ first: PAGE_LIMIT });
        const m = this.getOwnerComponent().getModel("todos");
        m.setProperty("/paging", {
          mode: "first",
          limit: PAGE_LIMIT,
          after: null,
          before: null,
        });
      },

      onLogout: function () {
        window.location.replace("http://localhost:3434/logout");
      },

      /* =================== RANDOM ACTIVITY =================== */
      onSuggestPress: async function () {
        if (!this._oActivityModel) {
          this._oActivityModel = new JSONModel({
            loading: true,
            activity: "",
            type: "",
            participants: 0,
            kidFriendly: null,
          });
        }
        if (!this._oActivityDlg) {
          this._oActivityDlg = this._createActivityDialog();
        }
        await this._loadRandomActivity();
        this._oActivityDlg.open();
      },

      async _loadRandomActivity() {
        try {
          this._oActivityModel.setProperty("/loading", true);

          const payload = JSON.stringify({
            query: `query RandomActivity {
              result: randomActivity {
                activity
                type
                participants
                kidFriendly
              }
            }`,
          });

          const out = await GraphqlClient.fetch(payload);
          const a = out || {};
          this._oActivityModel.setData({
            loading: false,
            activity: a.activity || "—",
            type: a.type || "—",
            participants: a.participants ?? 0,
            kidFriendly: !!a.kidFriendly,
          });
        } catch (e) {
          console.error("Random activity failed", e);
          MessageToast.show("Could not fetch activity, try again.");
          this._oActivityModel.setData({
            loading: false,
            activity: "—",
            type: "—",
            participants: 0,
            kidFriendly: null,
          });
        }
      },

      _createActivityDialog() {
        const header = new Bar({
          contentLeft: [
            new Image({ src: "images/cute_gopher.png" }).addStyleClass(
              "dlgTitleIcon"
            ),
            new Title({
              text: "Need a break? 🎲 Random activity",
            }).addStyleClass("dlgTitleText"),
          ],
        });

        const title = new Title({
          text: "{act>/activity}",
          level: "H3",
        }).addStyleClass("sapUiMediumMarginBottom");

        const rowType = new HBox({
          items: [
            new Label({ text: "Type:", width: "7rem" }),
            new Text({ text: "{act>/type}" }),
          ],
        }).addStyleClass("sapUiTinyMarginBottom");

        const rowParticipants = new HBox({
          items: [
            new Label({ text: "Participants:", width: "7rem" }),
            new Text({ text: "{act>/participants}" }),
          ],
        }).addStyleClass("sapUiTinyMarginBottom");

        const rowKids = new HBox({
          items: [
            new Label({ text: "Kid-friendly:", width: "7rem" }),
            new Text({
              text: "{= ${act>/kidFriendly} === null ? '—' : (${act>/kidFriendly} ? 'Yes ✅' : 'No 🚫') }",
            }),
          ],
        });

        const loadingText = new Text({ text: "Loading…" }).addStyleClass(
          "sapUiSmallMarginBottom"
        );

        const box = new VBox({
          items: [loadingText, title, rowType, rowParticipants, rowKids],
        }).addStyleClass("sapUiSmallMargin");

        const dlg = new Dialog({
          customHeader: header,
          contentWidth: "420px",
          content: [box],
          beginButton: new Button({
            text: "Another one",
            icon: "sap-icon://refresh",
            type: "Emphasized",
            press: () => this._loadRandomActivity(),
          }),
          endButton: new Button({
            text: "Close",
            press: function () {
              dlg.close();
            },
          }),
          afterOpen: () => {
            // show/hide loading line reactively
            const isLoading = this._oActivityModel.getProperty("/loading");
            loadingText.setVisible(!!isLoading);
            this._oActivityModel.attachPropertyChange(function (e) {
              if (e.getParameter("path") === "/loading") {
                loadingText.setVisible(!!e.getParameter("value"));
              }
            });
          },
        });

        dlg.setModel(this._oActivityModel, "act");
        this.getView().addDependent(dlg);
        return dlg;
      },

      /* =================== CREATE / EDIT / INFO / DELETE =================== */
      _toRFC3339FromDatePicker(dp) {
        const d = dp.getDateValue();
        if (!d) return "";
        return new Date(
          Date.UTC(d.getFullYear(), d.getMonth(), d.getDate())
        ).toISOString();
      },

      onAddPress() {
        if (!this._oAddDlg) this._oAddDlg = this._createAddDialog();
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

        const oDlgModel = new JSONModel({ selectedAssigneeId: "" });

        const oAssignee = new ComboBox({
          placeholder: "Assignee (optional)",
          width: "100%",
          selectedKey: "{dlg>/selectedAssigneeId}",
          items: {
            path: "collaborators>/users",
            template: new Item({
              key: "{collaborators>id}",
              text: "{collaborators>email}",
            }),
          },
        }).addStyleClass("dlgInput");
        oAssignee.setModel(
          this.getOwnerComponent().getModel("collaborators"),
          "collaborators"
        );

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

        const oDlg = new Dialog({
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
            new Label({ text: "Assignee", labelFor: oAssignee }),
            oAssignee,
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
                const assigneeId = oDlgModel.getProperty("/selectedAssigneeId");
                if (assigneeId) input.assignedTo = assigneeId;

                const payload = JSON.stringify({
                  query: `mutation CreateTodo($input: CreateTodoInput!) {
                    result: createTodo(input: $input) {
                      id name description status priority dueDate
                      assignedTo { id email }
                    }
                  }`,
                  variables: { input },
                });

                const newTodo = await GraphqlClient.fetch(payload);
                if (!newTodo?.id) throw new Error("Create failed");

                await this._reloadCurrentPage();
                await this._refillPageToLimit();
                MessageToast.show(`Todo "${name}" created`);
                oDlg.close();
              } catch (err) {
                console.error(err);
                MessageToast.show("Error while creating todo");
              }
            },
          }).addStyleClass("dlgPrimaryBtn"),
          endButton: new Button({
            text: "Cancel",
            press: () => oDlg.close(),
          }).addStyleClass("dlgPrimaryBtn"),
          afterClose: () => {
            oName.setValue("");
            oDesc.setValue("");
            oPriortiy.setValue("");
            oDlgModel.setProperty("/selectedAssigneeId", "");
          },
        });

        oDlg.setModel(oDlgModel, "dlg");
        return oDlg;
      },

      onBackPress() {
        const appRoot = sap.ui.require.toUrl("ui5/walkthrough");
        window.location.replace(appRoot + "/index.html");
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

        const oDlgModel = new JSONModel({
          selectedAssigneeId: oTodo.assignedTo?.id || "",
        });

        const oAssigneeEdit = new ComboBox({
          placeholder: "Assignee (optional)",
          width: "100%",
          selectedKey: "{dlg>/selectedAssigneeId}",
          items: {
            path: "collaborators>/users",
            template: new Item({
              key: "{collaborators>id}",
              text: "{collaborators>email}",
            }),
          },
        }).addStyleClass("dlgInput");
        oAssigneeEdit.setModel(
          this.getOwnerComponent().getModel("collaborators"),
          "collaborators"
        );

        const oHeader = new Bar({
          contentLeft: [
            new Image({ src: "images/cute_gopher.png" }).addStyleClass(
              "dlgTitleIcon"
            ),
            new Title({ text: "Update Todo" }).addStyleClass("dlgTitleText"),
          ],
        });

        const oDlg = new Dialog({
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
              text: "New Status",
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
            new Label({
              text: "Assignee",
              labelFor: oAssigneeEdit,
              class: "sapUiTinyMarginTop",
            }),
            oAssigneeEdit,
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
                const input = { name, description: desc, priority, status };
                const dueISO = this._toRFC3339FromDatePicker(oDue);
                if (dueISO) input.dueDate = dueISO;

                const selectedId = oDlgModel.getProperty("/selectedAssigneeId");
                if (selectedId) input.assignedTo = selectedId;

                const payload = JSON.stringify({
                  query: `mutation UpdateTodo($id: ID!, $input: UpdateTodoInput!) {
                    result: updateTodo(id: $id, input: $input) {
                      id name description status priority dueDate
                      assignedTo { id email }
                    }
                  }`,
                  variables: { id: oTodo.id, input },
                });

                const updTodo = await GraphqlClient.fetch(payload);
                if (!updTodo?.id) throw new Error("Update failed");

                await this._reloadCurrentPage();
                await this._refillPageToLimit();
                MessageToast.show("Todo was successfully updated");
                oDlg.close();
              } catch (err) {
                console.error(err);
                MessageToast.show("Error while updating todo");
              }
            },
          }),
          endButton: new Button({ text: "Cancel", press: () => oDlg.close() }),
        });

        oDlg.setModel(oDlgModel, "dlg");
        return oDlg;
      },

      async onInfoPress(oEvent) {
        const oButton = oEvent.getSource();
        const oCtx = oButton.getBindingContext("todos");
        if (!oCtx) {
          MessageToast.show("No row context");
          return;
        }
        const todoId = oCtx.getProperty("id");

        if (!this._oInfoPopover) {
          this._oInfoPopover = new Popover({
            placement: sap.m.PlacementType.Bottom,
            showHeader: false,
            contentWidth: "280px",
            content: new VBox({
              renderType: "Bare",
              items: [
                new HBox({
                  items: [
                    new Label({ text: "Created:", width: "7rem" }),
                    new Text({ text: "{/createdAt}" }),
                  ],
                }),
                new HBox({
                  items: [
                    new Label({ text: "Updated:", width: "7rem" }),
                    new Text({ text: "{/lastUpdated}" }),
                  ],
                }),
                new HBox({
                  items: [
                    new Label({ text: "Assignee:", width: "7rem" }),
                    new Text({ text: "{/assignedTo}" }),
                  ],
                }),
              ],
            }),
          });
          this.getView().addDependent(this._oInfoPopover);
        }

        const oModel = new JSONModel({
          createdAt: "Loading…",
          lastUpdated: "Loading…",
          assignedTo: "Loading…",
        });
        this._oInfoPopover.setModel(oModel);
        this._oInfoPopover.openBy(oButton);

        try {
          const payload = JSON.stringify({
            query: `query GetTodoInfo($id: ID!) {
              result: todo(id: $id) {
                createdAt
                lastUpdated
                assignedTo { email }
              }
            }`,
            variables: { id: todoId },
          });

          const info = (await GraphqlClient.fetch(payload)) || {};
          const fmt = sap.ui.core.format.DateFormat.getDateTimeInstance({
            style: "medium",
          });

          oModel.setData({
            createdAt: info.createdAt
              ? fmt.format(new Date(info.createdAt))
              : "—",
            lastUpdated: info.lastUpdated
              ? fmt.format(new Date(info.lastUpdated))
              : "—",
            assignedTo: info.assignedTo?.email || "—",
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

              const m = this.getOwnerComponent().getModel("todos");
              const pi = m.getProperty("/pageInfo") || {};
              const itemsBefore = m.getProperty("/todos") || [];

              if (itemsBefore.length <= 1) {
                if (pi.hasPrevPage) {
                  await this._fetchTodosPage({
                    last: PAGE_LIMIT,
                    before: pi.startCursor,
                  });
                  m.setProperty("/paging", {
                    mode: "prev",
                    limit: PAGE_LIMIT,
                    before: pi.startCursor,
                  });
                } else {
                  await this._fetchTodosPage({ first: PAGE_LIMIT });
                  m.setProperty("/paging", {
                    mode: "first",
                    limit: PAGE_LIMIT,
                    after: null,
                    before: null,
                  });
                }
                await this._refillPageToLimit();
              } else {
                await this._reloadCurrentPage();
                await this._refillPageToLimit();
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
