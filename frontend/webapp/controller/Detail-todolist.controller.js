sap.ui.define([
    "sap/ui/core/mvc/Controller",
    "sap/ui/core/routing/History",
    "sap/m/MessageToast",
    "sap/ui/model/json/JSONModel",
	"ui5/walkthrough/services/APIService"
], (Controller, History, MessageToast, JSONModel, ApiHelper) => {
    "use strict";

    return Controller.extend("ui5.walkthrough.controller.Detail", {

        onInit() {
            const oViewModel = new JSONModel({});
            this.getView().setModel(oViewModel, "view");

            const oRouter = this.getOwnerComponent().getRouter();
            oRouter.getRoute("detail-todolist").attachPatternMatched(this.onObjectMatched, this);
        },

        onObjectMatched(oEvent) {
            const todoListId = parseInt(oEvent.getParameter("arguments").todoListPath, 10);
            const todoLists = this.getOwnerComponent().getModel("todoLists").getData();
            const sTodoListPath = window.decodeURIComponent(oEvent.getParameter("arguments").todoListPath);
            const todoList = todoLists[todoListId];

            this.getView().bindElement({
                path: "/" + sTodoListPath,
                model: "todoLists"
            });

            this.getView().setModel(new JSONModel({ todoListId: todoList.id }), "view");
            this.fetchTodosData(todoList.id);
        },

        async fetchTodosData(todoListId) {
            try {
                const todosData = await ApiHelper.makeAPICall(`/todo-lists/get-todos/${todoListId}`);

                const oTodosModel = new JSONModel(todosData);
                this.getView().setModel(oTodosModel, "todos");

            } catch (error) {
                MessageToast.show("Error fetching todos data: " + error.message);
            }
        },

        onDeleteTodoFromTodoList: async function (oEvent) {
            const oButton = oEvent.getSource();

            const oBindingContext = oButton.getParent().getBindingContext("todos");
        
            const sTodoId = oBindingContext.getProperty("id");

            const todoListId = this.getView().getModel("view").getProperty("/todoListId");

            try {
                const options = {
                    method: "DELETE"
                };

                await ApiHelper.makeAPICall(`/todo-lists/${todoListId}/remove-todo/${sTodoId}`, options);

                MessageToast.show(`Todo with id ${sTodoId} deleted from todolist with id ${todoListId}`);

                this.fetchTodosData(todoListId); 

            } catch (error) {
                console.error("Error deleting todo:", error);
                MessageToast.show("Error deleting todo");
            }
        },



        onNavBack() {
            const oHistory = History.getInstance();
            const sPreviousHash = oHistory.getPreviousHash();

            if (sPreviousHash !== undefined) {
                window.history.go(-1);
            } else {
                const oRouter = this.getOwnerComponent().getRouter();
                oRouter.navTo("overview", {}, true);
            }
        }
    });
});
