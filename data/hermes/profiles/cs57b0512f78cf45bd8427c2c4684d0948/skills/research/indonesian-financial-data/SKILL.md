---
name: indonesian-financial-data
description: "Indonesian bank account, e-wallet, BPJS verification APIs."
tags: [indonesia, osint, financial, bank-account, api, verification]
---

# Indonesian Financial Data & Verification Tools

## Overview

Indonesian public financial verification APIs (bank account name lookup, e-wallet validation, BPJS checks) landscape as of 2024-2026.

## Current State (Updated Aug 2026)

### Dead Public APIs ❌

All free public APIs for bank account validation are **dead or expired**:

- `api-rekening.my.id` - connection refused
- `netovas.com/api/cekrek` - returns "Not Found"
- `api-rekening.lfourr.com` - domain for sale
- `makira.id` - API unavailable

### Working Commercial Options ✅

1. **OneBrick API** (docs.onebrick.io)
   - Indonesian bank account validation
   - Requires registration

2. **Xendit** 
   - Bank Account Validation API
   - Requires business account

3. **RapidAPI: cek-rekening-indonesia**
   - Requires API key (paid)

## GitHub Repositories

Most Indonesian "cek rekening" repos on GitHub are **wrappers around dead APIs**, not self-contained solutions.

### Best Self-Hostable Option

**RomySaputraSihananda/cek-rekening**
- Repo: https://github.com/RomySaputraSihananda/cek-rekening
- Stack: Node.js, TypeScript, Express, Swagger
- Status: ⚠️ Source code complete, but depends on dead `api-rekening.lfourr.com`
- To use: clone, replace API endpoint in `src/utils/Search.ts` with working alternative

### Other Repos (Reference Only)

- `haxsinner/cekrekening` - dead API (api-rekening.my.id)
- `kodingkeun/cekrekening.github.io` - dead API (netovas.com)
- `pilarxyz/cek-rekening` - dead API (lfourr.com)
- `makiraid/cek-norek-py` - dead API (makira.id)
- `nexcloudsss/NEOSINT-TRACKER` - promotional README only, no source

## Pitfalls

- **Don't assume GitHub stars/recent updates mean working tool** - most are abandoned wrappers
- **Public gratisan APIs don't survive long** - domains expire, maintenance stops
- **Self-hosting requires replacing dead upstream APIs** - check source before assuming it works

## Recommendation

For production Indonesian bank validation:
1. Use commercial APIs (OneBrick, Xendit) if budget allows
2. For research/OSINT: use RomySaputraSihananda repo as base, implement own scraping/validation logic
3. Never rely on free public Indonesian financial APIs for critical workflows
