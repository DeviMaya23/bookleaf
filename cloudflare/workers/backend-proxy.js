const CLOUD_RUN_URL = "https://bookleaf-backend-258487485284.australia-southeast1.run.app";

export default {
  async fetch(request) {
    const url = new URL(request.url);
    url.hostname = new URL(CLOUD_RUN_URL).hostname;

    const proxiedRequest = new Request(url, {
      method: request.method,
      headers: request.headers,
      body: request.body,
      redirect: "follow",
    });

    return fetch(proxiedRequest);
  },
};
