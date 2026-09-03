---
name: temporary-sms-otp-verification
description: "Use when sourcing virtual numbers for SMS OTP verification."
version: 1.0.0
platforms: [linux, darwin, windows]
metadata:
  hermes:
    tags: [sms, otp, virtual-number, 5sim, smspool, grizzlysms, sms-activate, pva, verification, cheap-numbers]
---

# Temporary SMS OTP & Virtual Number Verification

Use when procuring temporary virtual phone numbers for SMS OTP verification across global services (e.g. Jio, WhatsApp, Telegram, Google, Discord, TikTok, banking/shopping apps).

## 1. Provider Comparison & Price Routing

When recommending or automating virtual numbers, select the provider based on target service, budget, and carrier requirements:

| Provider | Best For / Key Strength | Typical Pricing Range | Minimum Deposit | Notes / Refund Policy |
| :--- | :--- | :--- | :--- | :--- |
| **5SIM** (`5sim.net`) | Bulk volume, rock-bottom pricing ($0.02 - $0.15), fast API | $0.02 - $0.10 for most Asian services | ~$1 - $2 | Auto-refund to balance if SMS not received within 15 mins. Large stock for India/Indonesia/SEA. |
| **Grizzly SMS** (`grizzlysms.com`) | Anti-VoIP / Real Carrier routes (Airtel, Vi, Jio) | $0.10 - $0.30 | ~$1.50 | Real SIM lines; higher success rate on fraud-sensitive platforms (Jio, OTTs). |
| **SMS-Activate** (`sms-activate.org`) | Broadest global coverage, long-term rentals & activation | $0.10 - $0.50 | Variable | Auto-refund on timeout. Supports both single OTP and rental pools. |
| **SMS-Man** (`sms-man.com`) | Budget alternative, multi-currency / crypto friendly | $0.15 - $0.35 | ~$1 - $2 | Clean API, good fallback if 5sim has low stock. |
| **SMSPool** (`smspool.net`) | US/EU non-VoIP quality verification | $0.50 - $3.50+ | $3.00 | **Overpriced for unlisted Asian services** ($3.00+ for "Not Listed"). Strict 14-day refund to payment source only if 100% untouched. |

---

## 2. API Probing & Price Checks

### 5SIM Price & Stock Lookup (Public Endpoint, No Auth Required)
```python
import urllib.request, json

def check_5sim(country="india", service="any"):
    url = f"https://5sim.net/v1/guest/products/{country}/{service}"
    req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
    return json.loads(urllib.request.urlopen(req, timeout=10).read().decode("utf-8"))

# Example: Get JioMart price in India
data = check_5sim("india", "any")
print(data.get("jiomart")) # {'Category': 'activation', 'Qty': 120000+, 'Price': 0.05}
```

### SMSPool Price Check (Public Endpoint)
```python
import urllib.request, json

def check_smspool(country_id=15, service_id=817):
    url = f"https://api.smspool.net/request/price?country={country_id}&service={service_id}"
    req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
    return json.loads(urllib.request.urlopen(req, timeout=10).read().decode("utf-8"))
```

---

## 3. Critical Pitfalls & Rules

- **Beware of "Unlisted / Not Listed" Pricing on SMSPool:** On SMSPool, requesting an unlisted/general service (service ID `817`) often incurs premium pricing ($3.00+) compared to specialized Russian/European aggregators (5SIM/Grizzly/SMS-Activate at $0.05 - $0.20). Always verify specific catalog prices before recommending high-rate platforms to budget-conscious users ($0.01 - $0.40 range).
- **Deposit Refund Conditions:** Most SMS verification platforms **do not support direct bank transfers / withdrawals of partially spent deposits**. Funds are committed as credit balance. Full refunds (e.g. on SMSPool) strictly require 0% spend within 14 days and return only to original payment methods (Stripe/Crypto).
- **Non-VoIP & Geo-IP Matching:** Platforms like Jio/Telecom providers detect virtual VoIP and proxy mismatches. When verifying OTP, ensure the user or automation uses a residential VPN/proxy matching the country code (+91 for India) to avoid server-side fraud blocks.
