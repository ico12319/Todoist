sap.ui.define([

], function(){
    "use strict";

    return {
      getCookieValue: function (cookieName) {
        return (
          document.cookie
            .match("(^|;)\\s*" + cookieName + "\\s*=\\s*([^;]+)")
            ?.pop() || ""
        );
      },

      hasCookie: function (cookieName) {
        return document.cookie
          .split(";")
          .some((cookie) => cookie.trim().startsWith(cookieName + "="));
      },

      setCookie: function (name, value, options = {}) {
        let cookieStr =
          encodeURIComponent(name) + "=" + encodeURIComponent(value);

        if (options.maxAge != null) {
          cookieStr += "; Max-Age=" + options.maxAge;
        }
        if (options.expires instanceof Date) {
          cookieStr += "; Expires=" + options.expires.toUTCString();
        }
        if (options.path) {
          cookieStr += "; Path=" + options.path;
        }
        if (options.domain) {
          cookieStr += "; Domain=" + options.domain;
        }
        if (options.secure) {
          cookieStr += "; Secure";
        }
        if (options.sameSite) {
          cookieStr += "; SameSite=" + options.sameSite;
        }

        document.cookie = cookieStr;
      },
    };
})