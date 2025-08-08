sap.ui.define([

], function(){
    "use strict";

    return Controller.extend("ui5.walkthrough.controller.Logout", {
      onLogout: function () {
        window.location.href = "http://localhost:3434/logout";
      },
    });

});