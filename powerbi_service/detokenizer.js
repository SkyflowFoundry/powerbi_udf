const { app } = require("@azure/functions");
const axios = require("axios");
const { Client } = require("@microsoft/microsoft-graph-client");
require("isomorphic-fetch");
const { generateBearerTokenFromCreds } = require("skyflow-node");
const qs = require("qs"); // helps build x-www-form-urlencoded content

async function getAppOnlyToken(tenantId, clientId, clientSecret) {
  const tokenEndpoint = `https://login.microsoftonline.com/${tenantId}/oauth2/v2.0/token`;

  const data = qs.stringify({
    grant_type: "client_credentials",
    client_id: clientId,
    client_secret: clientSecret,
    scope: "https://graph.microsoft.com/.default",
  });

  try {
    const response = await axios.post(tokenEndpoint, data, {
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
    });

    return response.data.access_token;
  } catch (error) {
    console.error(
      "Error getting token:",
      error.response?.data || error.message
    );
    throw error;
  }
}

app.http("detokenize", {
  methods: ["GET", "POST"],
  authLevel: "anonymous",
  handler: async (request, context) => {
    try {
      context.log(`Http function processed request for url "${request.url}"`);
      const skyflowRoles = [
        {
          name: "Global Administrator",
          ID: "skyflowRoleID",
        },
        { name: "Attribute Log Reader", ID: "skyflowRoleID" },
      ];
      const queryParams = request.query;
      const tokens = queryParams.getAll("token");
      let roleID = "";
      const accessToken = await getAppOnlyToken(
        "tenant_id",
        "client_id",
        "client_secret"
      );
      if (!accessToken) {
        return {
          status: 401,
          body: "Missing AAD token",
          headers: {
            "Content-Type": "application/json",
          },
        };
      }
      const graphClient = Client.init({
        authProvider: (done) =>
          done(
            null,
            accessToken
          ),
      });
      const userId = request.headers.get("X-MS-CLIENT-PRINCIPAL-NAME");
      let groups = [];
      groups = await graphClient.api(`/users/${userId}/memberOf`).get();
      const roles = groups.value.filter(
        (item) => item["@odata.type"] === "#microsoft.graph.directoryRole"
      );

      roleID = skyflowRoles.find((skyflowRole) => {
        const match = roles.find(
          (role) => skyflowRole.name == role.displayName
        );
        return match ? true : false;
      });
      const at = await generateBearerTokenFromCreds(
        JSON.stringify({
          // Service account credentials here
          // This can be fetched from function ENV variables 
        }),
        { roleIDs: [roleID.ID] }
      );
      const detokenizeRequest = tokens.map((token) => {
        return {
          "token": token
        }
      })
      // Can be fetched from ENV config or set in the function itself
      const detokenized = await axios.post(
        `${VAULT_URL}/v1/vaults/${VAULT_ID}/detokenize`,
        {
          detokenizationParameters: detokenizeRequest
        },
        {
          headers: {
            Authorization: `Bearer ${at.accessToken}`,
            "Content-Type": "application/json",
          },
        }
      );
      const detokenizeResponse = detokenized.data["records"].map(element => {
       return {
          token: element["token"],
          value: element["value"]
        }
      });
      return {
        status: 200,
        body: JSON.stringify(detokenizeResponse),
        headers: {
          "Content-Type": "application/json",
        },
      };
    } catch (error) {
      context.log("Error during detokenize API call:", error?.data ?? error?.message);
      return {
        status: 500,
        body: JSON.stringify({
          value: "ERROR",
          message: error?.data ?? error?.message,
        }),
        headers: {
          "Content-Type": "application/json",
        },
      };
    }
  },
});
