sap.ui.define([
	"sap/ui/core/mvc/Controller",
	"sap/m/MessageToast",
	"sap/ui/model/Filter",
	"sap/ui/model/FilterOperator",
	"ui5/walkthrough/services/APIService",
    "ui5/walkthrough/services/GraphqlClient"
], (Controller, MessageToast, Filter, FilterOperator, ApiHelper, graphqlUtils ) => {
	"use strict";

	return Controller.extend("ui5.walkthrough.controller.TodosList", {
		onInit: function () {
		},

		onPress(oEvent) {
			const oItem = oEvent.getSource();
			const oRouter = this.getOwnerComponent().getRouter();
			oRouter.navTo("detail", {
				todosPath: window.encodeURIComponent(oItem.getBindingContext("todo").getPath().substring(1))
			});
		},

		onSearchTodos(oEvent) {
			const aFilter = [];
			const sQuery = oEvent.getParameter("query");
		
			if (sQuery) {
				aFilter.push(new Filter("description", FilterOperator.Contains, sQuery));
			}
		
			const oList = this.byId("todoList");
			const oBinding = oList.getBinding("items");
		
			oBinding.filter(aFilter);
		},

		onDelete: async function (oEvent) {
            const oItem = oEvent.getSource().getParent();
            const oBindingContext = oItem.getBindingContext("todo"); 
            const sTodoId = oBindingContext.getProperty("id");

            try {
                const options = {
                    method: "DELETE"
                };

                await ApiHelper.makeAPICall(`/todos/${sTodoId}`, options);

                MessageToast.show("Todo deleted successfully");

                this._refreshTodos(); 

            } catch (error) {
                console.error("Error deleting todo:", error);
                MessageToast.show("Error deleting todo");
            }
        },

		_refreshTodos: async function () {
            const options = { method: "GET" };

            try {
                const data = await ApiHelper.makeAPICall("/todos", options);
                const oModel = new sap.ui.model.json.JSONModel(data);
				this.getView().setModel(oModel, "todo");
                this.getOwnerComponent().setModel(oModel, "todo");
            } catch (error) {
                console.error("Error fetching data:", error);
            }
        },

		onOpenDialog: function () {
            const oView = this.getView();
            const oDialog = oView.byId("addTodoDialog");

            if (!oDialog) {
                sap.ui.xmlfragment(oView.getId(), "ui5.walkthrough.view.AddTodoDialog", this);
                oView.addDependent(oDialog);
            }

            oDialog.open();
        },

        onDialogClose: function () {
            const oDialog = this.getView().byId("addTodoDialog");
            oDialog.close();
        },

		isFormValid: function () {
			const oNameInput = this.byId("todoNameInput");
			const oDescriptionInput = this.byId("todoDescriptionInput");
		
			const isNameValid = oNameInput.getValue().trim() !== "";
			const isDescriptionValid = oDescriptionInput.getValue().trim() !== "";
		
			return isNameValid && isDescriptionValid;
		},

		onInputChange: function () {
            this._checkFormValidity();
        },

		_checkFormValidity: function () {
			console.log('vliza')
            const oSaveButton = this.byId("saveButton");
			console.log(this.isFormValid())
            oSaveButton.setEnabled(this.isFormValid());
        },
		
		onSaveTodo: async function () {
            const oNameInput = this.byId("todoNameInput");
            const oDescriptionInput = this.byId("todoDescriptionInput");

            const newTodo = {
                name: oNameInput.getValue(),
                description: oDescriptionInput.getValue()
            };

            const options = {
                method: "POST",
                body: newTodo
            };

            try {
                const data = await ApiHelper.makeGraphQLCall(graphqlUtils.createTodoMutation, newTodo);
                console.log(data)
                MessageToast.show("Todo added successfully");
                this.onDialogClose();

                this._refreshTodos();
            } catch (error) {
                console.error("Error saving todo:", error);
                MessageToast.show("Error saving todo");
            }
        }
	});
});
