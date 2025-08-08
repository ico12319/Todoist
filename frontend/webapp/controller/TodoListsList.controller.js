sap.ui.define([
	"sap/ui/core/mvc/Controller",
	"sap/m/MessageToast",
	"sap/ui/model/Filter",
	"sap/ui/model/FilterOperator",
	"ui5/walkthrough/services/APIService"
], (Controller, MessageToast, Filter, FilterOperator, ApiHelper ) => {
	"use strict";

	return Controller.extend("ui5.walkthrough.controller.TodoListsList", {
		onInit: function () {
		},

		onPressAction(oEvent) {
			const oItem = oEvent.getSource();
			const oRouter = this.getOwnerComponent().getRouter();
			oRouter.navTo("detail-todolist", {
				todoListPath: window.encodeURIComponent(oItem.getBindingContext("todoLists").getPath().substring(1))
			});
		},

		onSearchTodoLists(oEvent) {
			const aFilter = [];
			const sQuery = oEvent.getParameter("query");
		
			if (sQuery) {
				aFilter.push(new Filter("name", FilterOperator.Contains, sQuery));
			}
		
			const oList = this.byId("todoLists");
			const oBinding = oList.getBinding("items");
		
			oBinding.filter(aFilter);
		},

		onAddTodoToList() {
			const todoIdInput = this.byId("todoIdInput").getValue();
		
			if (!todoIdInput) {
				MessageToast.show("Please enter a Todo ID.");
				return;
			}
		
			const selectedListId = this._selectedListId;
		
		
			const options = {
				method: "POST",
				headers: {
					"Content-Type": "application/json"
				}
			};
		
			ApiHelper.makeAPICall(`/todo-lists/${selectedListId}/add-todo/${todoIdInput}`, options)
				.then(() => {
					MessageToast.show(`Todo with ID ${todoIdInput} added to list with ID ${selectedListId}`);
					this.onDialogClose();
				})
				.catch(error => {
					console.error("Error adding todo:", error);
					MessageToast.show("Error adding todo");
					this.onDialogClose();
				})
		},

		onOpenDialog: function () {
            const oView = this.getView();
            const oDialog = oView.byId("addTodoToTodoListDialog");

            if (!oDialog) {
                sap.ui.xmlfragment(oView.getId(), "ui5.walkthrough.view.AddTodoDialog", this);
                oView.addDependent(oDialog);
            }

            oDialog.open();
        },

		onOpenDialog: function (oEvent) {
			var oClickedItem = oEvent.getSource().getParent();
		
			var oContext = oClickedItem.getBindingContext("todoLists");
			var listId = oContext.getProperty("id");
			this._selectedListId = listId;

			const oView = this.getView();
            const oDialog = oView.byId("addTodoToTodoListDialog");
			console.log(this._selectedListId)
            if (!oDialog) {
                sap.ui.xmlfragment(oView.getId(), "ui5.walkthrough.view.AddTodoDialog", this);
                oView.addDependent(oDialog);
            }

            oDialog.open();
		},

		onDialogClose: function () {
            const oDialog = this.getView().byId("addTodoToTodoListDialog");
            oDialog.close();
        },
	});
});
