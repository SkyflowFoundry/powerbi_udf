# Power BI Azure App Integration for Detokenization

## Overview

Power BI Service runs under a **service identity**, which means it does not carry the logged-in user’s context during data refresh or transformation. This poses a challenge when trying to perform **detokenization** — a security-sensitive operation that should be controlled by **user-specific access policies**, such as those defined by **Azure AD groups**.

To address this, a secure and scalable integration was designed using **Azure infrastructure** that enables **role-specific detokenization** of data columns in Power BI, while maintaining compliance and security via **Microsoft Entra** (formerly Azure AD).

---

## High-Level Architecture

![Power BI Service detokenization architecture](../images/PowerBIService%3C%3ESkyflow.svg)

1. **Power BI Service (Power Query)** initiates a call to a protected **Azure Function** endpoint to fetch detokenized data.

2. Since **Power Query** does not hold the actual user identity, it authenticates to the Azure Function using an **organizational account (OAuth2)**.

3. The **Azure Function App**, secured with **Microsoft Entra ID**, extracts the calling user’s identity from built-in headers:
   - `x-ms-client-principal`
   - `x-ms-client-principal-name`

4. The Azure Function then:
   - Calls the **Microsoft Graph API** to retrieve the user’s **Azure AD groups**
   - Uses these group details to **authorize** and perform **detokenization** by calling a downstream detokenization API/service

5. The **detokenized data** is returned to Power BI via Power Query.

---

## Authentication & Authorization Design

- A dedicated **Azure App Registration** was created to represent the integration between Power BI and the Azure Function.

- The App was configured with **Microsoft Entra authentication**, allowing only **organizational users** to access the Azure Function.

- A **custom domain** was used as the **Application URI** to meet Microsoft Entra’s requirement for **verified domains** in the `aud` (audience) claim of tokens.

- By validating the **user identity** and fetching **Azure AD groups** in real-time, the function could enforce **fine-grained access control** for detokenization.

---

## Power BI Integration Approach

- While Power BI doesn’t support **dynamic, user-level transformations** out of the box, **Power Query** was used to **outsource** the transformation to an **external service (Azure Function)**.

- Within Power Query, a **secured HTTP call** was made to the Azure Function using the **Organizational Account** authentication type.

---

## Limitations

- When using Organizational Account as the authentication method, the logged-in user’s identity can be retrieved, but the request cannot include a payload. Only GET requests are supported, so any tokens must be passed through query parameters.

- This may introduce performance overhead.

## Outcome

This architecture ensures that:

- Detokenization occurs **only if the logged-in user is authorized** (based on Azure AD groups)
- Power BI can display **sensitive data securely**, respecting **role-based policies**
- The approach is **extensible** to multiple datasets, users, and roles across the organization
