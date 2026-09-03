# Indian SMS OTP Verification & Jio Ecosystem Details

## Jio Verification Carrier Architecture
- Jio services (MyJio, JioMart, JioCinema, JioFiber) perform carrier-type checks and rate limiting on incoming registrations.
- Pure VoIP ranges (e.g. standard Twilio, TextNow, Google Voice) get immediately rejected or fail to receive OTP SMS.
- Real carrier routing (non-VoIP / physical SIM pools from Airtel, Reliance Jio, and Vodafone Idea / Vi) is required.

## Live Public Price Probes

### 1. 5SIM (`https://5sim.net/v1/guest/products/india/any`)
- `jiomart`: ~$0.05
- `adani`, `apple`, `line`, `signal`, `my11circle`: ~$0.04
- `aliexpress`: ~$0.02
- Auto-refund timer: 15 minutes if no SMS received.

### 2. Grizzly SMS (`https://grizzlysms.com/myjio`)
- Real carrier routing specifically indexed for `myjio` and `jiomart`.
- Average price: $0.10 - $0.20 per activation.

### 3. SMSPool (`https://api.smspool.net/request/price?country=15&service=817`)
- Country 15 (India), Service 817 (Not Listed / Other): ~$3.01.
- High cost due to unlisted fallback surcharge. Avoid recommending SMSPool for budget Indian SMS verification.

## Payment & Refund Matrix
- **SMSPool**: Full refund to Stripe/Crypto if 0% of deposit spent within 14 days. Once partially used, balance is locked to the account.
- **5SIM / SMS-Activate**: Low minimum deposit ($1 - $2), failed OTP auto-credits back to account balance within 15 mins.
