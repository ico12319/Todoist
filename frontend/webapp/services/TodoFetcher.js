sap.ui.define(
  [
    "sap/ui/model/json/JSONModel",
    "sap/m/MessageToast",
    "ui5/walkthrough/services/GraphqlClient",
  ],
  function (JSONModel, MessageToast, GraphqlClient) {
    "use strict";

    return {
      fetchNextPageTodos: async function (oController, listId, first, after) {
        const payload = JSON.stringify({
          query: `query GetTodos($id: ID!, $first: Int, $after: ID) {
      result: list(id: $id) {
        todos(first: $first, after: $after){
          data { 
            id 
            name 
            description 
            status 
            priority
            dueData
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
          const list = await GraphqlClient.fetch(payload); // list
          const node = list?.todos || {};
          const aTodos = node.data || [];
          pageInfo = node.pageInfo || {};
          const totalCount = node.totalCount ?? 0;

          const oComponent = oController.getOwnerComponent();
          let oModel = oComponent.getModel("todos");
          if (!oModel) {
            oModel = new JSONModel({
              todos: [],
              pageInfo: {},
              hasNext: false,
              hasPrev: false,
              totalCount: 0,
            });
            oComponent.setModel(oModel, "todos");
          }

          oModel.setData({
            todos: aTodos,
            pageInfo,
            hasNext: !!pageInfo.hasNextPage,
            hasPrev: !!pageInfo.hasPrevPage,
            totalCount,
          });

          MessageToast.show(`Loaded ${aTodos.length} todos`);
        } catch (err) {
          console.error("Error fetching todos (next):", err);
          MessageToast.show("Error loading next page");
        }
        return pageInfo;
      },

      fetchPrevPageTodos: async function (oController, listId, last, before) {
        const payload = JSON.stringify({
          query: `query GetTodos($id: ID!, $last: Int, $before: ID) {
            result: list(id: $id) {
                    todos(last: $last, before: $before){
                     data { 
                        id 
                        name 
                        description 
                        status 
                        priority 
                        dueData
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
          variables: { id: listId, last: last, before: before },
        });

        let pageInfo = {};
        try {
          const list = await GraphqlClient.fetch(payload);
          const node = list?.todos || {};
          const aTodos = node.data || [];
          pageInfo = node.pageInfo || {};
          const totalCount = node.totalCount ?? 0;

          const oComponent = oController.getOwnerComponent();
          let oModel = oComponent.getModel("todos");
          if (!oModel) {
            oModel = new JSONModel({
              todos: [],
              pageInfo: {},
              hasNext: false,
              hasPrev: false,
              totalCount: 0,
            });
            oComponent.setModel(oModel, "todos");
          }

          oModel.setData({
            todos: aTodos,
            pageInfo,
            hasNext: !!pageInfo.hasNextPage,
            hasPrev: !!pageInfo.hasPrevPage,
            totalCount,
          });

          MessageToast.show(`Loaded ${aTodos.length} todos`);
        } catch (err) {
          console.error("Error fetching todos (prev):", err);
          MessageToast.show("Error loading previous page");
        }
        return pageInfo;
      },
    };
  }
);
