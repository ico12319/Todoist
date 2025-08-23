sap.ui.define(["ui5/walkthrough/util/Cookie"], function (Cookie) {
  "use strict";

  async function refreshAccessToken() {
    const refreshToken = Cookie.getCookieValue("refresh-token");
    if (!refreshToken) throw new Error("missing refresh token");

    const mutation = `
    mutation ($input: RefreshTokenInput!) {
      result: exchangeRefreshToken(input: $input) {
        jwtToken
        refreshToken
      }
    }`;

    const res = await fetch("http://graphql-service:8090/query", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({
        query: mutation,
        variables: { input: { refreshToken } },
      }),
    });
    if (!res.ok) throw new Error("Failed to refresh access token");

    const { data } = await res.json();
    const { jwtToken, refreshToken: newRefresh } = data.result;

    Cookie.setCookie("access-token", jwtToken, { path: "/", maxAge: 3600 });
    Cookie.setCookie("refresh-token", newRefresh, { path: "/", maxAge: 86400 });

    return jwtToken;
  }

  return{
    refreshAccessToken
  };

});
