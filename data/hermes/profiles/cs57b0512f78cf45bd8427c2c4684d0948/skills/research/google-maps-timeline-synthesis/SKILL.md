---
name: google-maps-timeline-synthesis
description: "Use when synthesizing Google Maps Timeline JSON exports."
tags: [google-maps, timeline, location-history, json, synthesis, osint, simulation]
---

# Google Maps Timeline (Location History) Synthesis & Editing

## Overview
Guidelines and generator patterns for synthesizing or editing realistic Google Maps Semantic Location History (`location-history.json`) exports without triggering anomaly detection or looking artificially generated.

## Google Semantic Location History Schema

Google Maps Timeline export JSON structure consists of an array of objects representing either a `visit`, a travel `activity`, or waypoint paths `timelinePath`:

### 1. Place Visit (`visit`)
```json
{
  "startTime": "2026-01-10T08:30:00.249+07:00",
  "endTime": "2026-01-11T15:34:18.127+07:00",
  "visit": {
    "hierarchyLevel": "0",
    "probability": "0.829898",
    "topCandidate": {
      "placeID": "ChIJt9RwegB5ei4RUFk8rZtXmXI",
      "placeLocation": "geo:-7.330619,110.508287",
      "probability": "0.270805",
      "semanticType": "Unknown"
    }
  }
}
```

### 2. Travel Activity (`activity`)
```json
{
  "startTime": "2026-01-22T09:10:19.159+07:00",
  "endTime": "2026-01-22T11:22:51.363+07:00",
  "activity": {
    "start": "geo:-7.330250,110.508556",
    "end": "geo:-7.814054,110.925323",
    "distanceMeters": "78685.842019",
    "probability": "0.883707",
    "topCandidate": {
      "type": "in passenger vehicle",
      "probability": "0.656923"
    }
  }
}
```
*Valid `topCandidate.type` values:* `in passenger vehicle`, `walking`, `in train`, `flying`, `on bicycle`, `running`, `still`.

### 3. Timeline Path Waypoints (`timelinePath`)
```json
{
  "startTime": "2026-01-22T08:00:00.000Z",
  "endTime": "2026-01-22T11:00:00.000Z",
  "timelinePath": [
    {
      "point": "geo:-7.492072,110.651392",
      "durationMinutesOffsetFromStartTime": "44"
    }
  ]
}
```
*Note: `timelinePath` often uses UTC timestamp strings (`.000Z`) rounded to hours, whereas `visit` and `activity` use local timezone offsets (e.g. `+07:00`, `+08:00`, `+09:00`).*

## Realistic Movement Synthesis Rules

To ensure generated location histories appear completely authentic:
1. **Local Micro-Movements & Day Trips:** Do not leave an agent static at a single coordinate for weeks. Add frequent local day trips (cafes, transit hubs, adjacent subdistricts/towns), short walks (15-95 meters, 1-3 minutes), and return-home legs.
2. **Coordinate Micro-Jitter:** Add random gaussian/uniform micro-jitter (`±0.0001` to `±0.0003` degrees) on repeatedly visited locations to mimic real GPS drift.
3. **Realistic Travel Physics:**
   - Road vehicles: 25-45 km/h urban, 60-85 km/h tollways.
   - Rail / Train: 60-90 km/h with station routes.
   - Commercial Flight: 450-750 km/h + 30-45 min ground/runway buffer.
4. **Timezone Transitions:** Ensure timestamps switch timezone offsets naturally when crossing time zones (e.g. WIB `+07:00` -> WITA/HKT `+08:00` -> JST `+09:00`).
5. **Floating-Point Probabilities:** Google formats probability and distance fields as stringified decimals with 6 fractional digits (e.g. `"0.883707"`, `"78685.842019"`).
