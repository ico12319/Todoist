sap.ui.define(
  [
    "sap/ui/model/json/JSONModel",
    "sap/m/MessageToast",
    "ui5/walkthrough/services/GraphqlClient",
  ],
  function (JSONModel, MessageToast, GraphqlClient) {
    "use strict";

    return {
      fetchNextPageLists: async function (oController, first, after) {
        const payload = JSON.stringify({
          query: `
          query GetLists($first: Int, $after: ID) {
            result: lists(first: $first, after: $after) {
              data {
                id
                name
                description
              }
              pageInfo {
                startCursor
                endCursor
                hasNextPage
                hasPrevPage
              }
              totalCount
            }
          }
        `,
          variables: { first: first, after: after },
        });

        let pageInfo = null;
        try {
          const listPage = await GraphqlClient.fetch(payload);
          const aLists = listPage?.data || [];
          pageInfo = listPage.pageInfo || null;

          const oComponent = oController.getOwnerComponent();
          let oModel = oComponent.getModel("lists");
          if (!oModel) {
            oModel = new JSONModel({
              lists: [],
              pageInfo: {},
              hasNext: false,
              hasPrev: false,
              totalCount: 0,
            });
            oComponent.setModel(oModel, "lists");
          }

          oModel.setData({
            lists: aLists,
            pageInfo: pageInfo,
            hasNext: pageInfo?.hasNextPage || false,
            hasPrev: pageInfo?.hasPrevPage || false,
            totalCount: listPage.totalCount,
          });

          MessageToast.show(`Заредени са ${aLists.length} списъка`);
        } catch (err) {
          console.error("Error fetching lists:", err);
          MessageToast.show("Грешка при зареждане на списъци");
        }

        return pageInfo;
      },

      fetchPrevPageLists: async function (oController, last, before) {
        const payload = JSON.stringify({
          query: `
          query GetLists($last: Int, $before: ID) {
            result: lists(last: $last, before: $before) {
              data {
                id
                name
                description
              }
              pageInfo {
                startCursor
                endCursor
                hasNextPage
                hasPrevPage
              }
              totalCount
            }
          }
        `,
          variables: { last: last, before: before },
        });

        let pageInfo = null;
        try {
          const listPage = await GraphqlClient.fetch(payload);
          const aLists = listPage?.data || [];
          pageInfo = listPage.pageInfo || null;

          const oComponent = oController.getOwnerComponent();
          let oModel = oComponent.getModel("lists");
          if (!oModel) {
            oModel = new JSONModel({
              lists: [],
              pageInfo: {},
              hasNext: false,
              hasPrev: false,
              totalCount: 0,
            });
            oComponent.setModel(oModel, "lists");
          }

          oModel.setData({
            lists: aLists,
            pageInfo: pageInfo,
            hasNext: pageInfo?.hasNextPage || false,
            hasPrev: pageInfo?.hasPrevPage || false,
            totalCount: listPage.totalCount,
          });

          MessageToast.show(`Заредени са ${aLists.length} списъка`);
        } catch (err) {
          console.error("Error fetching lists:", err);
          MessageToast.show("Грешка при зареждане на списъци");
        }

        return pageInfo;
      },
    };
  }
);
