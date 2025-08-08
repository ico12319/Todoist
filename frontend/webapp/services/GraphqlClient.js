sap.ui.define([
  "ui5/walkthrough/util/Cookie",
  "ui5/walkthrough/services/Refresh"
], function (Cookie, Refresh) {
  "use strict";

  return {
    fetch: async (query) => {
      let headers = {
        "Content-Type": "application/json",
        Accept: "application/json",
      };

      const accessToken = Cookie.getCookieValue("access-token");

      headers.Authorization = accessToken;

      let oResponse = await fetch("http://localhost:8090/query", {
        method: "POST",
        body: query,
        headers,
        credentials: "include",
      });

      let payload = await oResponse.json();

      const isUnauthorized = payload.errors?.some(
        (e) => e.extensions?.code == "UNAUTHORIZED"
      );

      if (isUnauthorized) {
        console.log("epa bate ne si otoriziran sori motori")
        await Refresh.refreshAccessToken();

        headers.Authorization = Cookie.getCookieValue("access-token");
       
        oResponse = await fetch("http://localhost:8090/query", {
          method: "POST",
          body: query,
          headers,
          credentials: "include",
        });

        payload = await oResponse.json();
      }


      if (payload.errors?.length) {
        throw new Error(payload.errors[0].message || "GraphQL error");
      }

      return payload.data?.result ?? null;
    },
  };
});
