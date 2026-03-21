function readCookie(name) {
  const prefix = name + "=";
  const parts = document.cookie ? document.cookie.split("; ") : [];

  for (const part of parts) {
    if (part.startsWith(prefix)) {
      return decodeURIComponent(part.slice(prefix.length));
    }
  }

  return "";
}

window.onload = function () {
  window.ui = SwaggerUIBundle({
    url: "/swagger/openapi.yaml",
    dom_id: "#swagger-ui",
    deepLinking: true,
    displayRequestDuration: true,
    persistAuthorization: true,
    tryItOutEnabled: true,
    presets: [
      SwaggerUIBundle.presets.apis
    ],
    layout: "BaseLayout",
    requestInterceptor: function (request) {
      request.credentials = "same-origin";

      const method = (request.method || "GET").toUpperCase();
      if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
        const csrfToken = readCookie("csrf_token");
        if (csrfToken) {
          request.headers["X-CSRF-Token"] = csrfToken;
        }
      }

      return request;
    }
  });
};
