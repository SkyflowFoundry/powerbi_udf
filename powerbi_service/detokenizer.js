/** Sample Azure Function entry point that performs role-aware Skyflow detokenization for Power BI. */
"use strict";

const { app } = require("@azure/functions");
const axios = require("axios");
const { Client } = require("@microsoft/microsoft-graph-client");
require("isomorphic-fetch");
const { generateBearerTokenFromCreds } = require("skyflow-node");
const qs = require("qs"); // build x-www-form-urlencoded payloads for AAD

const {
  SKYFLOW_SERVICE_ACCOUNT = "{}", // This can be fetched from Azure Key Vault or other secure location
  SKYFLOW_VAULT_URL: VAULT_URL,
  SKYFLOW_VAULT_ID: VAULT_ID,
} = process.env;

/** Map Azure AD directory roles to Skyflow role IDs that the service account may assume. */
const SKYFLOW_ROLE_MAPPINGS = [
  { name: "Azure Role1", id: "skyflowRoleID1" },
  { name: "Azure Role2", id: "skyflowRoleID2" },
];

async function getAppOnlyToken(tenantId, clientId, clientSecret) {
  const tokenEndpoint = `https://login.microsoftonline.com/${tenantId}/oauth2/v2.0/token`;

  const data = qs.stringify({
    grant_type: "client_credentials",
    client_id: clientId,
    client_secret: clientSecret,
    scope: "https://graph.microsoft.com/.default",
  });

  const response = await axios.post(tokenEndpoint, data, {
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
  });

  return response.data.access_token;
}

function findSkyflowRoleId(directoryRoles) {
  const match = SKYFLOW_ROLE_MAPPINGS.find(({ name }) =>
    directoryRoles.some((role) => role.displayName === name)
  );
  return match?.id;
}

function buildDetokenizationRequest(tokens) {
  return tokens.map((token) => ({ token }));
}

app.http("detokenize", {
  methods: ["GET", "POST"],
  authLevel: "anonymous",
  handler: async (request, context) => {
    try {
      context.log(`Http function processed request for url "${request.url}"`);

      if (!VAULT_URL || !VAULT_ID) {
        throw new Error("Missing VAULT_URL or VAULT_ID configuration.");
      }

      const tokens = request.query?.getAll("token") ?? [];
      if (!tokens.length) {
        return {
          status: 400,
          body: JSON.stringify({ message: "At least one token query parameter is required." }),
          headers: { "Content-Type": "application/json" },
        };
      }

      const userId = request.headers.get("X-MS-CLIENT-PRINCIPAL-NAME");
      if (!userId) {
        return {
          status: 401,
          body: JSON.stringify({ message: "Missing user identity header." }),
          headers: { "Content-Type": "application/json" },
        };
      }

      const accessToken = await getAppOnlyToken("tenant_id", "client_id", "client_secret");
      if (!accessToken) {
        return {
          status: 401,
          body: JSON.stringify({ message: "Unable to acquire Microsoft Graph token." }),
          headers: { "Content-Type": "application/json" },
        };
      }

      const graphClient = Client.init({
        authProvider: (done) => done(null, accessToken),
      });

      // Use Microsoft Graph to materialize the caller's directory roles.
      const membership = await graphClient.api(`/users/${encodeURIComponent(userId)}/memberOf`).get();
      const directoryRoles =
        membership.value?.filter(
          (item) => item["@odata.type"] === "#microsoft.graph.directoryRole"
        ) ?? [];

      const skyflowRoleId = findSkyflowRoleId(directoryRoles);
      if (!skyflowRoleId) {
        return {
          status: 403,
          body: JSON.stringify({ message: "User is not mapped to a Skyflow role." }),
          headers: { "Content-Type": "application/json" },
        };
      }

      const serviceAccountToken = await generateBearerTokenFromCreds(
        SKYFLOW_SERVICE_ACCOUNT,
        { roleIDs: [skyflowRoleId] }
      );

      const detokenizationPayload = {
        detokenizationParameters: buildDetokenizationRequest(tokens),
      };

      const detokenized = await axios.post(
        `${VAULT_URL}/v1/vaults/${VAULT_ID}/detokenize`,
        detokenizationPayload,
        {
          headers: {
            Authorization: `Bearer ${serviceAccountToken.accessToken}`,
            "Content-Type": "application/json",
          },
        }
      );

      const detokenizedRecords =
        detokenized.data.records?.map((record) => ({
          token: record.token,
          value: record.value,
        })) ?? [];

      return {
        status: 200,
        body: JSON.stringify(detokenizedRecords),
        headers: { "Content-Type": "application/json" },
      };
    } catch (error) {
      context.log("Error during detokenize API call:", error?.response?.data ?? error?.message);
      return {
        status: 500,
        body: JSON.stringify({
          value: "ERROR",
          message: error?.response?.data ?? error?.message,
        }),
        headers: { "Content-Type": "application/json" },
      };
    }
  },
});
