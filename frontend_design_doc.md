# 💻 Frontend Design Document --- Crypto Wallet & Transaction Aggregator

**Framework:** Next.js 15 (App Router) + TypeScript\
**Style System:** TailwindCSS + ShadCN/UI + Framer Motion\
**API Layer:** REST (connects to Go backend via `/api/v1/*`)

------------------------------------------------------------------------

## 🧭 1. Overall Goals

Frontend must act as: - Multi-chain Wallet Dashboard for BTC, ETH, SOL,
XLM - Transaction Manager to send/track status - Swap Engine UI
(off-chain swap mock) with live rates - KYC / Profile Center for
document verification - Audit & Notifications View for system activity -
Responsive / PWA-ready UI for Desktop, Tablet, POS devices

------------------------------------------------------------------------

## 🧩 2. Tech Stack Summary

  ------------------------------------------------------------------------
  Category              Stack / Library                Purpose
  --------------------- ------------------------------ -------------------
  **Framework**         Next.js 15 (App Router)        Full-stack React
                                                       framework

  **Language**          TypeScript                     Strongly-typed
                                                       development

  **UI Framework**      TailwindCSS + ShadCN/UI        Reusable component
                                                       system

  **Animation**         Framer Motion                  Smooth transitions

  **Charts**            Recharts                       For analytics and
                                                       visualizations

  **Form Handling**     React Hook Form + Zod          Input validation

  **State Management**  Zustand                        Lightweight global
                                                       store

  **API Client**        Axios or fetch wrapper         Communicate with
                                                       backend

  **Auth**              JWT (HttpOnly Cookie)          Authentication and
                                                       sessions

  **Intl / i18n**       next-intl                      Multilingual
                                                       (English / Thai)

  **Notifications**     React Hot Toast / ShadCN Toast Alerts and status
                                                       messages

  **Layout System**     Tailwind Grid + Flexbox        Dashboard layout
  ------------------------------------------------------------------------

------------------------------------------------------------------------

## 🧱 3. Folder Structure Example

``` bash
frontend/
├── app/
│   ├── layout.tsx
│   ├── page.tsx (Dashboard)
│   ├── wallets/
│   │   ├── page.tsx
│   │   ├── [id]/page.tsx
│   ├── transactions/
│   │   ├── page.tsx
│   │   └── [txid]/page.tsx
│   ├── swap/page.tsx
│   ├── kyc/page.tsx
│   ├── notifications/page.tsx
│   └── settings/page.tsx
│
├── components/
│   ├── ui/
│   ├── wallet/
│   ├── tx/
│   ├── swap/
│   ├── charts/
│   └── layout/
│
├── lib/
│   ├── api.ts
│   ├── auth.ts
│   ├── useWalletStore.ts
│   ├── constants.ts
│   └── helpers.ts
│
├── styles/
│   └── globals.css
│
└── package.json
```

------------------------------------------------------------------------

## 🧭 4. Key Pages & Components

### 🏠 Dashboard

-   Overview of total balance (multi-chain)
-   Recharts graph of trends
-   Quick actions: Send / Receive / Swap

### 💼 Wallets Page

-   List all user wallets
-   Show per-chain balance + address
-   Components:
    -   `WalletCard`
    -   `AddWalletModal`

### 💸 Transactions Page

-   History of transactions with status filter
-   `TxStatusBadge` component for visual cues

### 🔁 Swap Page

-   `SwapForm` for from/to tokens
-   `SwapRateCard` for live exchange rates
-   `SwapHistory` table for past swaps

### 🪪 KYC Page

-   Upload and manage identity documents
-   Display KYC status (pending, approved, rejected)

### 🔔 Notifications Page

-   Shows latest system events from audit logs

### ⚙️ Settings Page

-   Account preferences, 2FA, Theme, Language

------------------------------------------------------------------------

## 🎨 5. UI Design Language

  Concept          Description
  ---------------- ------------------------------------
  **Style**        Apple-like glassmorphism aesthetic
  **Colors**       Slate background + chain gradients
  **Icons**        Lucide React Icons
  **Animation**    Framer Motion transitions
  **Typography**   Inter / DM Sans / Prompt

------------------------------------------------------------------------

## 🔐 6. Auth & Session Flow

``` mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend (Next.js)
    participant B as Backend (Go API)

    U->>F: Login (email/password)
    F->>B: POST /api/v1/auth/login
    B-->>F: Set HttpOnly JWT cookie
    F->>B: GET /api/v1/wallets
    B-->>F: Return wallet data
    F->>U: Render dashboard
```

------------------------------------------------------------------------

## 🔗 7. API Connection Layer

``` ts
import axios from "axios";

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1",
  withCredentials: true,
});

export const getWallets = async () => (await api.get("/wallets")).data;
export const getRates = async () => (await api.get("/rates")).data;
export const postTransaction = async (payload: any) =>
  (await api.post("/transactions", payload)).data;
```

------------------------------------------------------------------------

## 📊 8. Analytics Widgets

  ---------------------------------------------------------------------------------------
  Widget                   Source                           Description
  ------------------------ -------------------------------- -----------------------------
  **RateChart**            `/api/v1/rates`                  Token price history

  **TxVolumeChart**        `/api/v1/transactions/summary`   Transaction volume

  **KYCCompletionChart**   `/api/v1/kyc/summary`            Verification progress

  **BalanceTrend**         `/api/v1/wallets/summary`        Total balance trend
  ---------------------------------------------------------------------------------------

------------------------------------------------------------------------

## 🪄 9. Developer Experience

  Tool                       Usage
  -------------------------- -------------------------
  **Hot Reload**             `npm run dev`
  **Type Safety**            TypeScript + Zod
  **Mock API**               Next.js Route Handlers
  **Storybook (optional)**   Component documentation
  **E2E Test**               Playwright

------------------------------------------------------------------------

## ✅ 10. Deliverables

1.  Folder structure (`app/`, `components/`, `lib/`)
2.  Routing setup (App Router)
3.  Tailwind + ShadCN theme with chain-specific gradients
4.  Zustand global store
5.  Auth-protected pages
6.  Wallet + Swap pages functional with backend mock
7.  Dashboard analytics widgets

------------------------------------------------------------------------

**Summary:**\
This frontend acts as the unified wallet dashboard, providing real-time
crypto management across Bitcoin, Ethereum, Solana, and Stellar. Built
for scalability, security, and clean UX design.

------------------------------------------------------------------------
