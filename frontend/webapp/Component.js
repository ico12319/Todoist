sap.ui.define([
	"sap/ui/core/UIComponent",
	"sap/ui/model/json/JSONModel",
	"sap/ui/Device"
], (UIComponent, JSONModel, Device) => {
	"use strict";

	return UIComponent.extend("ui5.walkthrough.Component", {
		metadata: {
			interfaces: ["sap.ui.core.IAsyncContentCreation"],
			manifest: "json"
		},

        init() {
            UIComponent.prototype.init.apply(this, arguments);

			const lists = {
				name: "",
				description: "",
			}


            const oModel = new JSONModel(lists);
            this.setModel(oModel, "lists");

            this.getRouter().initialize();
        },
	});
});
