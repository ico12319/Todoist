sap.ui.define([], function () {
    "use strict";
    const backendURL = "http://localhost:8090";
    
    return {
        /**
         * Generic function to make API calls
         * @param {string} path The API URL to call
         * @param {object} oOptions The options for the request (method, headers, body, etc.)
         * @returns {Promise} A promise resolving to the response data
         */
        makeAPICall: async function (path, oOptions) {
            const defaultOptions = {
                method: "GET",
                headers: {
                    "Content-Type": "application/json"
                },
                body: null
            };
        
            const options = Object.assign({}, defaultOptions, oOptions);
        
            const accessToken = localStorage.getItem("accessToken");
            if (accessToken) {
                options.headers["Authorization"] = `${accessToken}`;
            }
        
            if (options.body && typeof options.body !== 'string') {
                options.body = JSON.stringify(options.body);
            }
        
            try {
                const response = await fetch(backendURL + path, options);
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }
        
                if (response.status === 204) {
                    return null;
                }

                if (response.status === 200 && !response.headers.get("content-type").includes("json")) {
                    return null;
                }
        
                const data = await response.json();
                return data;
            } catch (error) {
                console.error("Error during API request:", error);
                throw error;
            }
        },

        /**
         * Function to make a GraphQL API call
         * @param {string} query The GraphQL query or mutation
         * @param {object} variables The variables for the GraphQL request (optional)
         * @returns {Promise} A promise resolving to the response data
         */
        makeGraphQLCall: async function (query, variables = {}) {
            const graphQLQuery = {
                query: query,
                variables: variables
            };

            const options = {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify(graphQLQuery)
            };

            const accessToken = localStorage.getItem("accessToken");
            if (accessToken) {
                options.headers["Authorization"] = `${accessToken}`;
            }

            try {
                const response = await fetch(backendURL + '/query', options);
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }

                const data = await response.json();
                if (data.errors) {
                    throw new Error(`GraphQL error: ${JSON.stringify(data.errors)}`);
                }

                return data.data;
            } catch (error) {
                console.error("Error during GraphQL request:", error);
                throw error;
            }
        }
    };
});
