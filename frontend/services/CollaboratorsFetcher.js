sap.ui.define(
  [
    "sap/ui/model/json/JSONModel",
    "sap/m/MessageToast",
    "ui5/walkthrough/services/GraphqlClient",
  ],
  function (JSONModel, MessageToast, GraphqlClient) {
    "use strict";

    return {
      fetchNextPageCollaborators: async function (
        oController,
        listId,
        first,
        after
      ) {
        const payload = JSON.stringify({
          query: `query GetCollaborators($id: ID!, $first: Int, $after: ID) {
      result: list(id: $id) {
        collaborators(first: $first, after: $after){
          data { 
            id 
            email
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
          variables: { id: listId, first, after },
        });

        let pageInfo = {};
        try {
          const list = await GraphqlClient.fetch(payload);
          const node = list?.collaborators || {};
          const aCollaborators = node.data || [];
          pageInfo = node.pageInfo || {};
          const totalCount = node.totalCount ?? 0;

          const oComponent = oController.getOwnerComponent();
          let oModel = oComponent.getModel("collaborators");
          if (!oModel) {
            oModel = new JSONModel({
              users: [],
              pageInfo: {},
              hasNext: false,
              hasPrev: false,
              totalCount: 0,
            });
            oComponent.setModel(oModel, "collaborators");
          }

          oModel.setData({
            users: aCollaborators,
            pageInfo,
            hasNext: !!pageInfo.hasNextPage,
            hasPrev: !!pageInfo.hasPrevPage,
            totalCount,
          });

          MessageToast.show(`Loaded ${aCollaborators.length} collaborators`);
        } catch (err) {
          console.error("Error fetching collaborators (next):", err);
          MessageToast.show("Error loading next page");
        }
        return pageInfo;
      },

      fetchPrevPageCollaborators: async function (
        oController,
        listId,
        last,
        before
      ) {
        const payload = JSON.stringify({
          query: `query GetCollaborators($id: ID!, $last: Int, $before: ID) {
      result: list(id: $id) {
        collaborators(last: $last, before: $before){
          data { id email }
          pageInfo { startCursor endCursor hasPrevPage hasNextPage }
          totalCount
        }
      }
    }`,
          variables: { id: listId, last, before },
        });

        let pageInfo = {};
        try {
          const list = await GraphqlClient.fetch(payload); 
          const node = list?.collaborators || {};
          const aCollaborators = node.data || [];
          pageInfo = node.pageInfo || {};
          const totalCount = node.totalCount ?? 0;

          const oComponent = oController.getOwnerComponent();
          let oModel = oComponent.getModel("collaborators");
          if (!oModel) {
            oModel = new JSONModel({
              users: [],
              pageInfo: {},
              hasNext: false,
              hasPrev: false,
              totalCount: 0,
            });
            oComponent.setModel(oModel, "collaborators");
          }

          oModel.setData({
            users: aCollaborators,
            pageInfo,
            hasNext: !!pageInfo.hasNextPage,
            hasPrev: !!pageInfo.hasPrevPage,
            totalCount,
          });

          MessageToast.show(`Loaded ${aCollaborators.length} collaborators`);
        } catch (err) {
          console.error("Error fetching collaborators (prev):", err);
          MessageToast.show("Error loading previous page");
        }
        return pageInfo;
      },
    };
  }
);
