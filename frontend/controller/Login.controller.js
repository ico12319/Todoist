sap.ui.define([
    "sap/ui/core/mvc/Controller"
], function (Controller) {
    "use strict";

    return Controller.extend("ui5.walkthrough.controller.Login", {
        onLoginWithGitHub: function () {
            window.location.href = "http://localhost:3434/github/login";
        }
    });
});
