sap.ui.define(
  [
    "sap/ui/core/mvc/Controller",
    "ui5/walkthrough/services/ListFetcher",
    "ui5/walkthrough/util/Cookie",
  ],
  (
    Controller,
    ListFetcher,
    Cookie
  ) => {
    "use strict";

    return Controller.extend("ui5.walkthrough.controller.App", {
      onInit: function () {
        this.checkLoginStatus();
      },

      checkLoginStatus: function () {
        const accessToken = Cookie.getCookieValue("access-token");
        console.log(accessToken);

        if (!accessToken) {
          this.getOwnerComponent().getRouter().navTo("login");
        } else {
          ListFetcher.fetchNextPageLists(this, 5, null);
        }
      },

      onLogout: function () {
        window.location.replace("http://localhost:3434/logout");
      },

      onNavigateTodos: function () {
        this.getOwnerComponent().getRouter().navTo("overview");
      },

      onNavigateTodoListsList: function () {
        this.getOwnerComponent().getRouter().navTo("todolists");
      },
    });
  }
);
