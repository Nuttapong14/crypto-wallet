# Feature Specification: Multi-Chain Crypto Wallet and Exchange Platform

**Feature Branch**: `001-build-a-web`
**Created**: 2025-10-14
**Status**: Draft
**Input**: User description: "Build a Web Application by see a spec follow @tech_stack.md @system_architecture_full.md @realtime_rates_websocket_design.md"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Wallet Creation and Balance Viewing (Priority: P1)

A user wants to securely store and manage multiple cryptocurrencies (Bitcoin, Ethereum, Solana, Stellar) in one place, with the ability to view their current balances and track their portfolio value in real-time.

**Why this priority**: This is the core value proposition - users need to store crypto before they can do anything else. Without wallets and balance visibility, no other features are usable.

**Independent Test**: Can be fully tested by creating a user account, generating wallets for supported chains, and displaying balances with live price conversions to fiat currency.

**Acceptance Scenarios**:

1. **Given** a new user completes registration, **When** they access the dashboard, **Then** they are prompted to create wallets for supported blockchains
2. **Given** a user has created wallets, **When** they view their dashboard, **Then** they see all wallet addresses and current balances for each cryptocurrency
3. **Given** a user has crypto in their wallets, **When** prices update in real-time, **Then** their portfolio value automatically reflects current market rates
4. **Given** a user views their balance, **When** they request to see wallet addresses, **Then** they can copy addresses to receive funds

---

### User Story 2 - Send and Receive Cryptocurrency (Priority: P1)

A user wants to send cryptocurrency to other wallet addresses and receive crypto from others, with clear transaction status tracking and confirmations.

**Why this priority**: Transferring funds is a fundamental wallet capability - users must be able to move their assets. This is essential MVP functionality alongside wallet creation.

**Independent Test**: Can be fully tested by initiating a send transaction, tracking its status through confirmation, and receiving funds from an external wallet with proper balance updates.

**Acceptance Scenarios**:

1. **Given** a user has sufficient balance, **When** they initiate a send transaction with valid recipient address and amount, **Then** the transaction is queued and they receive a transaction ID
2. **Given** a transaction is submitted, **When** it's being processed by the blockchain, **Then** the user sees real-time status updates (pending → confirmed)
3. **Given** a transaction completes, **When** the user views transaction history, **Then** they see the completed transaction with timestamp, amount, fees, and blockchain confirmation
4. **Given** someone sends crypto to a user's wallet, **When** the transaction is confirmed, **Then** the balance automatically updates and the user receives a notification

---

### User Story 3 - Real-Time Price Monitoring (Priority: P2)

A user wants to monitor cryptocurrency prices in real-time to make informed decisions about when to buy, sell, or exchange their assets.

**Why this priority**: Price awareness is critical for decision-making but users can still store and transfer crypto without it. This enhances the core wallet functionality.

**Independent Test**: Can be fully tested by opening the dashboard and observing live price updates for BTC, ETH, SOL, and XLM without page refresh, with price changes reflected within 5 seconds of market movements.

**Acceptance Scenarios**:

1. **Given** a user opens the dashboard, **When** cryptocurrency prices change in the market, **Then** displayed prices update automatically without page reload
2. **Given** a user views price information, **When** prices update, **Then** they see visual indicators for price increases (green) and decreases (red)
3. **Given** a user has multiple cryptocurrencies, **When** viewing their portfolio, **Then** total portfolio value updates in real-time as prices fluctuate
4. **Given** a user experiences connection issues, **When** the connection is restored, **Then** prices sync to current market values automatically

---

### User Story 4 - Cryptocurrency Exchange (Priority: P2)

A user wants to exchange one cryptocurrency for another without leaving the platform, seeing fair exchange rates and estimated fees before confirming the swap.

**Why this priority**: Swapping between cryptocurrencies adds significant value but requires the foundational wallet and transaction infrastructure. Users can manage single-currency holdings without this.

**Independent Test**: Can be fully tested by selecting two cryptocurrencies, viewing the current exchange rate, executing a swap, and verifying that both wallet balances update correctly with transaction history recorded.

**Acceptance Scenarios**:

1. **Given** a user has sufficient balance in one cryptocurrency, **When** they request an exchange quote, **Then** they see the current exchange rate, estimated fees, and expected output amount
2. **Given** a user confirms a swap, **When** the exchange is processed, **Then** the source wallet decreases and destination wallet increases by the correct amounts
3. **Given** a swap is in progress, **When** the user checks transaction status, **Then** they see real-time progress updates
4. **Given** exchange rates fluctuate significantly, **When** a user has a pending quote, **Then** the quote expires after a reasonable timeframe (30-60 seconds) to protect against unfavorable execution

---

### User Story 5 - Transaction History and Analytics (Priority: P3)

A user wants to review their complete transaction history across all cryptocurrencies, with filtering, search, and export capabilities for personal record-keeping and tax purposes.

**Why this priority**: Historical analysis is valuable but not required for core wallet operations. Users can function with current state visibility while this provides enhanced tracking.

**Independent Test**: Can be fully tested by performing multiple transactions, viewing the history page, filtering by cryptocurrency/date/type, searching for specific transactions, and exporting data to CSV format.

**Acceptance Scenarios**:

1. **Given** a user has transaction history, **When** they access the history page, **Then** they see all transactions sorted by most recent with type, amount, date, status, and fees
2. **Given** a user views transaction history, **When** they apply filters (date range, cryptocurrency, transaction type), **Then** only matching transactions are displayed
3. **Given** a user needs records, **When** they request export, **Then** they receive a downloadable file with complete transaction details
4. **Given** a user has many transactions, **When** they scroll through history, **Then** transactions load progressively without performance degradation

---

### User Story 6 - Account Security and KYC Verification (Priority: P2)

A user wants to secure their account with strong authentication and complete identity verification to access higher transaction limits and ensure regulatory compliance.

**Why this priority**: Security is critical for trust but basic wallet functionality can work with standard authentication. KYC enables higher limits and compliance, supporting growth scenarios.

**Independent Test**: Can be fully tested by registering with email/password, enabling two-factor authentication, uploading identity documents, and receiving verification status updates with appropriate access level changes.

**Acceptance Scenarios**:

1. **Given** a new user registers, **When** they complete initial authentication setup, **Then** they can access basic wallet features with standard transaction limits
2. **Given** a user wants higher limits, **When** they initiate KYC verification and upload required documents, **Then** they receive clear status updates on verification progress
3. **Given** a user completes KYC verification, **When** verification is approved, **Then** their transaction limits increase and they receive confirmation notification
4. **Given** a user has security concerns, **When** they enable additional security measures (2FA, withdrawal whitelist), **Then** these protections are enforced on all sensitive operations

---

### User Story 7 - Portfolio Analytics Dashboard (Priority: P3)

A user wants to visualize their portfolio performance over time, including asset allocation, historical value trends, and gain/loss calculations.

**Why this priority**: Advanced analytics enhance user experience but aren't required for basic wallet operations. This is a value-add feature for engaged users.

**Independent Test**: Can be fully tested by building a portfolio over time, viewing performance charts, checking asset allocation pie charts, and verifying gain/loss calculations against actual transaction history.

**Acceptance Scenarios**:

1. **Given** a user has transaction history, **When** they access the analytics dashboard, **Then** they see charts showing portfolio value over time
2. **Given** a user holds multiple cryptocurrencies, **When** they view allocation, **Then** they see a visual breakdown of holdings by percentage and value
3. **Given** a user wants performance insights, **When** they view analytics, **Then** they see total gain/loss, best performing asset, and comparison to initial investment
4. **Given** a user wants detailed period analysis, **When** they select date ranges, **Then** charts and calculations adjust to show performance for that specific period

---

### Edge Cases

- **What happens when a user enters an invalid recipient address?** System validates the address format for the selected blockchain and shows a clear error message before allowing transaction submission
- **How does the system handle blockchain network congestion?** Users see estimated confirmation times based on current network conditions and can optionally increase transaction fees for faster processing
- **What happens when a user's session expires during a transaction?** Pending transactions are saved and users can resume or cancel upon re-authentication, with no loss of funds or data
- **How does the system handle concurrent balance updates?** All balance updates are processed atomically with proper locking to prevent race conditions or incorrect balance calculations
- **What happens when real-time price feeds fail?** System falls back to cached recent prices with a clear indicator that prices may be stale, and automatically reconnects when service is restored
- **How does the system handle insufficient gas fees for Ethereum transactions?** System estimates gas requirements before submission and warns users if their balance is insufficient to cover both the transfer amount and fees
- **What happens when KYC documents are rejected?** Users receive specific feedback on what needs to be corrected with the ability to re-submit updated documents
- **How does the system handle blockchain reorganizations (reorgs)?** System monitors confirmation depth and only marks transactions as fully confirmed after sufficient blockchain confirmations (6+ for Bitcoin)
- **What happens when a user tries to exchange but liquidity is insufficient?** System shows "insufficient liquidity" message and suggests alternative trading pairs or reduced amounts
- **How does the system prevent users from accidentally sending large amounts?** System implements confirmation steps for transactions above configurable thresholds and displays warnings for unusually large transfers

## Requirements *(mandatory)*

### Functional Requirements

**Wallet Management**:
- **FR-001**: System MUST allow users to create secure wallets for Bitcoin, Ethereum, Solana, and Stellar blockchains
- **FR-002**: System MUST generate unique, blockchain-specific wallet addresses for each supported cryptocurrency
- **FR-003**: System MUST display current balances for all user wallets with real-time updates
- **FR-004**: System MUST encrypt and securely store private keys with recovery mechanisms
- **FR-005**: System MUST allow users to export wallet addresses for receiving funds
- **FR-006**: System MUST support multiple wallets per user for each blockchain

**Transaction Management**:
- **FR-007**: System MUST enable users to send cryptocurrency by specifying recipient address and amount
- **FR-008**: System MUST validate recipient addresses against blockchain-specific format rules before submission
- **FR-009**: System MUST estimate and display transaction fees before users confirm transactions
- **FR-010**: System MUST track transaction status from submission through blockchain confirmation
- **FR-011**: System MUST update wallet balances immediately when transactions are confirmed on the blockchain
- **FR-012**: System MUST provide transaction receipts with transaction hash, timestamp, amount, and fees
- **FR-013**: System MUST queue transactions during high-load periods and process them in order
- **FR-014**: System MUST detect and handle failed transactions with clear error messages and refund mechanisms

**Price Information**:
- **FR-015**: System MUST display real-time cryptocurrency prices for BTC, ETH, SOL, and XLM
- **FR-016**: System MUST update displayed prices automatically without requiring page refresh
- **FR-017**: System MUST show 24-hour price changes with visual indicators (percentage and direction)
- **FR-018**: System MUST calculate and display total portfolio value in user's preferred fiat currency
- **FR-019**: System MUST maintain price update frequency of at least once every 5 seconds

**Cryptocurrency Exchange**:
- **FR-020**: System MUST allow users to exchange between supported cryptocurrencies
- **FR-021**: System MUST display current exchange rates with estimated output amounts before swap confirmation
- **FR-022**: System MUST show estimated fees and final amounts for all exchange operations
- **FR-023**: System MUST execute swaps atomically (both wallets update or neither updates)
- **FR-024**: System MUST record swap operations in transaction history with exchange rate and fees
- **FR-025**: System MUST expire exchange rate quotes after a reasonable timeframe to prevent stale pricing
- **FR-026**: System MUST handle partial order fills and notify users when liquidity is insufficient

**User Authentication & Security**:
- **FR-027**: System MUST require users to create accounts with secure authentication
- **FR-028**: System MUST enforce strong password requirements (minimum length, complexity)
- **FR-029**: System MUST support two-factor authentication for enhanced account security
- **FR-030**: System MUST implement session management with automatic timeout after inactivity
- **FR-031**: System MUST log all security-relevant events (logins, failed attempts, setting changes)
- **FR-032**: System MUST allow users to set withdrawal address whitelists for additional security

**KYC & Compliance**:
- **FR-033**: System MUST collect user identity information for regulatory compliance
- **FR-034**: System MUST allow users to upload identity verification documents
- **FR-035**: System MUST assign verification status levels (unverified, pending, verified)
- **FR-036**: System MUST enforce transaction limits based on verification status
- **FR-037**: System MUST provide clear feedback on verification status and requirements
- **FR-038**: System MUST securely store and encrypt personally identifiable information
- **FR-039**: System MUST implement data retention policies compliant with financial regulations

**Transaction History & Reporting**:
- **FR-040**: System MUST maintain complete transaction history for all user operations
- **FR-041**: System MUST allow users to filter transactions by date range, cryptocurrency, and type
- **FR-042**: System MUST provide search functionality for finding specific transactions
- **FR-043**: System MUST allow users to export transaction history in standard formats (CSV, PDF)
- **FR-044**: System MUST display transaction details including type, amount, fees, status, and blockchain confirmations
- **FR-045**: System MUST support pagination or infinite scroll for large transaction histories

**Analytics & Portfolio Management**:
- **FR-046**: System MUST calculate and display total portfolio value across all cryptocurrencies
- **FR-047**: System MUST show asset allocation breakdown by cryptocurrency and percentage
- **FR-048**: System MUST calculate profit/loss for each holding based on transaction history
- **FR-049**: System MUST display portfolio performance over selectable time periods
- **FR-050**: System MUST provide visual charts for portfolio value trends and asset allocation

**Notifications**:
- **FR-051**: System MUST notify users when transactions are confirmed
- **FR-052**: System MUST alert users when they receive incoming transfers
- **FR-053**: System MUST send notifications for security events (new login, settings changes)
- **FR-054**: System MUST provide KYC status update notifications
- **FR-055**: System MUST allow users to configure notification preferences (email, in-app, push)

**System Operations**:
- **FR-056**: System MUST maintain audit logs of all critical operations for compliance and debugging
- **FR-057**: System MUST implement rate limiting to prevent abuse and ensure fair usage
- **FR-058**: System MUST handle blockchain network downtime gracefully with appropriate user messaging
- **FR-059**: System MUST implement backup and recovery mechanisms for user data
- **FR-060**: System MUST monitor system health and alert administrators of critical issues

### Key Entities

- **User**: Represents a registered account holder with authentication credentials, verification status, preferences, and associated wallets. Contains profile information, security settings, and notification preferences.

- **Wallet**: Represents a blockchain-specific cryptocurrency wallet with unique address, encrypted private key, current balance, and transaction history. Associated with one user and one blockchain network.

- **Transaction**: Represents a cryptocurrency transfer operation including sender/receiver addresses, amount, fees, status (pending/confirmed/failed), blockchain transaction hash, timestamp, and confirmation count. Links to source and destination wallets.

- **Exchange Operation**: Represents a cryptocurrency swap between two assets including source/destination currencies, amounts, exchange rate, timestamp, and transaction references for both sides of the swap.

- **Price Data**: Represents current and historical cryptocurrency prices including symbol, price in various fiat currencies, 24-hour change, volume, and timestamp. Used for real-time displays and portfolio calculations.

- **KYC Profile**: Represents user identity verification information including verification level, submitted documents, verification status, and compliance notes. Encrypted and stored separately from operational data.

- **Notification**: Represents user notifications including type (transaction, security, system), message content, delivery status, timestamp, and user read status.

- **Audit Log**: Represents system events for compliance and security including action type, user identifier, timestamp, IP address, and action details. Immutable and stored separately.

- **Account Balance**: Represents aggregated balance information across all wallets including total portfolio value, per-cryptocurrency balances, and historical value for analytics.

## Success Criteria *(mandatory)*

### Measurable Outcomes

**User Experience**:
- **SC-001**: Users can create a new account and generate their first wallet in under 3 minutes
- **SC-002**: Users can complete a cryptocurrency send transaction in under 90 seconds from initiation to confirmation receipt
- **SC-003**: 95% of users successfully complete their first transaction without support assistance
- **SC-004**: Dashboard displays real-time price updates within 5 seconds of market price changes
- **SC-005**: 90% of new users complete at least one transaction within their first session

**Performance & Reliability**:
- **SC-006**: System supports 10,000 concurrent active users without performance degradation
- **SC-007**: Wallet balance displays load within 2 seconds of dashboard access
- **SC-008**: Transaction status updates appear within 10 seconds of blockchain confirmation
- **SC-009**: System maintains 99.9% uptime for core wallet and transaction functionality
- **SC-010**: Price data streams remain connected with less than 0.1% packet loss during normal operations

**Security & Compliance**:
- **SC-011**: Zero unauthorized access incidents to user wallets or private keys
- **SC-012**: 100% of security-critical events are logged in immutable audit trails
- **SC-013**: KYC document processing completes within 24 hours for 95% of submissions
- **SC-014**: All stored private keys remain encrypted with no plaintext exposure
- **SC-015**: System passes security audit with no critical or high-severity vulnerabilities

**Transaction Accuracy**:
- **SC-016**: 100% of successfully submitted transactions result in correct balance updates
- **SC-017**: Exchange operations execute with actual rates within 1% of quoted rates
- **SC-018**: Zero double-spend or duplicate transaction incidents
- **SC-019**: Transaction fee estimates are accurate within 10% of actual network fees
- **SC-020**: Portfolio value calculations are accurate within 0.5% of actual market values

**Business Metrics**:
- **SC-021**: User portfolio value tracking leads to 40% increase in platform engagement time
- **SC-022**: Real-time price displays result in 25% increase in exchange transaction frequency
- **SC-023**: Successful transaction rate exceeds 98% (excluding user errors and insufficient funds)
- **SC-024**: Average time to complete KYC verification is under 4 hours for standard cases
- **SC-025**: Transaction history export feature is used by at least 30% of active users monthly

## Assumptions

1. **Regulatory Compliance**: The platform will operate in jurisdictions requiring KYC verification and financial transaction monitoring. Specific regulatory requirements will be determined by target markets.

2. **Blockchain Network Access**: The system will use publicly accessible blockchain nodes or third-party node services (not self-hosted nodes initially). Node availability and reliability are assumed to be managed by service providers.

3. **Price Data Sources**: Cryptocurrency prices will be sourced from established market data providers (CoinGecko, Binance API, etc.) with sufficient API rate limits and reliability for real-time updates.

4. **User Base**: Initial target users are individuals with basic cryptocurrency knowledge who want to manage multiple blockchain assets in one place. Enterprise or institutional users are not the primary focus for MVP.

5. **Supported Cryptocurrencies**: The system will support four major blockchains (Bitcoin, Ethereum, Solana, Stellar) with their native currencies. ERC-20 tokens and other token standards can be added in future iterations.

6. **Transaction Volume**: Expected transaction volume is moderate (under 1,000 transactions per hour) for initial launch. The architecture should support scaling to higher volumes.

7. **Fiat Currency Support**: Portfolio values and prices will be displayed in major fiat currencies (USD, EUR) with conversion rates from the same price data sources as cryptocurrency prices.

8. **Customer Support**: Basic customer support will be available via email and in-app help documentation. Real-time chat support is not required for MVP.

9. **Mobile Access**: The web application will be mobile-responsive for access from smartphones and tablets. Native mobile apps are not included in initial scope.

10. **Transaction Finality**: Different blockchains have different finality times. The system will use industry-standard confirmation thresholds (e.g., 6 confirmations for Bitcoin, 12+ for Ethereum) before marking transactions as final.

11. **Data Retention**: Transaction and audit data will be retained for a minimum of 7 years to comply with standard financial regulations. User-initiated account deletion will follow GDPR-compliant data removal processes.

12. **Exchange Functionality**: The initial exchange feature will focus on off-chain order matching and balance updates. Integration with decentralized exchanges (DEXs) or advanced order types can be added later.

## Security & Privacy Considerations

### Security Requirements

**Authentication & Access Control**:
- Multi-factor authentication strongly recommended and required for accounts with high transaction volumes or balances
- Session tokens expire after 30 minutes of inactivity
- IP-based anomaly detection for unusual login patterns
- Account lockout after 5 failed login attempts with graduated unlock times

**Key Management**:
- Private keys encrypted at rest using industry-standard encryption algorithms
- Keys never transmitted in plaintext or stored in logs
- Hardware security module (HSM) or equivalent key management service for master encryption keys
- User passwords never stored in plaintext (hashed with strong, salted algorithms)

**Transaction Security**:
- All transaction requests require re-authentication for amounts above configurable thresholds
- Withdrawal address whitelisting option for additional security layer
- Mandatory waiting period (24-48 hours) for first-time withdrawal addresses
- Transaction signing occurs in secure, isolated environments

**Data Protection**:
- All data transmitted over encrypted connections (TLS 1.3+)
- Personally identifiable information (PII) encrypted at rest
- KYC documents stored in separate, encrypted storage with restricted access
- Database encryption for sensitive data columns

**Monitoring & Incident Response**:
- Real-time monitoring for suspicious transaction patterns
- Automated alerts for unusual account activity
- Incident response procedures for security events
- Regular security audits and penetration testing

### Privacy Requirements

**Data Minimization**:
- Only collect user data necessary for core functionality and regulatory compliance
- Allow users to view what data is stored about them
- Provide data export functionality for user-requested information

**Regulatory Compliance**:
- GDPR compliance for European users (data access, deletion, portability rights)
- Financial privacy regulations compliance based on operating jurisdictions
- Anti-money laundering (AML) monitoring and reporting capabilities
- Know Your Customer (KYC) verification with tiered access levels

**User Transparency**:
- Clear privacy policy explaining data collection and usage
- Explicit consent for optional data collection
- Notification of privacy policy changes
- Option to delete account and associated data (subject to legal retention requirements)

**Third-Party Integrations**:
- Documented data sharing with third-party services (KYC providers, blockchain node services, price data APIs)
- Contractual agreements with third parties for data protection
- Regular audits of third-party security practices

## Out of Scope (for initial version)

**Features Not Included**:
1. **Fiat On/Off Ramps**: Direct bank transfers, credit card purchases, or selling crypto for fiat currency
2. **Advanced Trading Features**: Limit orders, stop-loss, margin trading, or futures
3. **DeFi Integration**: Direct interaction with decentralized finance protocols, lending, staking, or yield farming
4. **NFT Support**: Non-fungible token storage, display, or trading
5. **Token Standards**: ERC-20, BEP-20, or other token standards beyond native blockchain currencies
6. **Multi-Signature Wallets**: Wallets requiring multiple approvals for transactions
7. **Cold Storage Integration**: Hardware wallet connectivity or cold storage management
8. **Social Features**: User-to-user chat, social trading, leaderboards, or public profiles
9. **Automated Trading**: Bots, algorithmic trading, or scheduled transactions
10. **Tax Reporting**: Automated tax calculation, capital gains reporting, or IRS form generation (beyond basic transaction export)
11. **Cross-Chain Bridges**: Direct token transfers between different blockchains
12. **Advanced Analytics**: Portfolio optimization suggestions, risk scoring, or investment recommendations
13. **Loyalty Programs**: Rewards, referral bonuses, or transaction fee discounts
14. **White-Label Solutions**: Multi-tenant architecture or partner platform integrations
15. **API Access**: Public APIs for external application integration or algorithmic trading

**Technical Scope Boundaries**:
- Desktop applications (native Windows/Mac/Linux apps)
- Native mobile applications (iOS/Android apps)
- Browser extensions or wallet plugins
- Blockchain node hosting or validator services
- Custom blockchain or Layer-2 solutions

**Compliance Scope Boundaries**:
- Money transmitter licensing (will use third-party licensed partners if required)
- Securities offering or trading (only supports cryptocurrency, not tokenized securities)
- Banking services or FDIC insurance

## Dependencies

**External Services**:
- Cryptocurrency price data API provider (CoinGecko, Binance, or equivalent)
- Blockchain node access services for BTC, ETH, SOL, XLM networks
- KYC verification service provider (Onfido, SumSub, or equivalent)
- Email delivery service for notifications and account verification
- SMS service for two-factor authentication (optional but recommended)
- Cloud infrastructure provider for hosting and storage

**Regulatory**:
- Legal review of terms of service and privacy policy
- Compliance framework determination based on operating jurisdictions
- Anti-money laundering (AML) and KYC requirement specifications
- Data protection regulation compliance (GDPR, CCPA, etc.)

**Technical Prerequisites**:
- Domain name and SSL certificates for secure web access
- Database infrastructure capable of handling financial transaction requirements
- Backup and disaster recovery infrastructure
- Monitoring and alerting infrastructure
- Development, staging, and production environment separation

**Team Dependencies**:
- Security expert review of key management and encryption implementation
- Blockchain developers familiar with BTC, ETH, SOL, XLM protocols
- Frontend developers experienced with real-time data display
- Backend developers with fintech or financial systems experience
- QA team with security testing and financial transaction testing expertise

## Open Questions

The following questions will be addressed through `/speckit.clarify`:

1. **Transaction Limits & Verification Tiers**: What specific transaction limits should apply to each KYC verification level (unverified, basic verified, fully verified)? What defines each verification tier?

2. **Exchange Liquidity Model**: How will the exchange functionality handle liquidity? Will it be purely internal order matching between users, integration with external exchanges, or a liquidity pool model?

3. **Supported Fiat Currencies**: Which fiat currencies should be supported for portfolio value display and price conversion? Should conversion be limited to USD and EUR, or include other major currencies (GBP, JPY, etc.)?
