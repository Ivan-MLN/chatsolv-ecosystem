#!/usr/bin/env python3
"""
Template generator for realistic Google Maps Timeline JSON data.
"""
import json
import random
import math
from datetime import datetime, timedelta

def haversine_distance(lat1, lon1, lat2, lon2):
    R = 6371000
    phi1, phi2 = math.radians(lat1), math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlam = math.radians(lon2 - lon1)
    a = math.sin(dphi / 2.0)**2 + math.cos(phi1)*math.cos(phi2)*math.sin(dlam / 2.0)**2
    return 2 * R * math.atan2(math.sqrt(a), math.sqrt(1 - a))

def jitter(coord, delta=0.00025):
    return coord + random.uniform(-delta, delta)

def generate_timeline_sample():
    # Sample usage demonstrating visit and activity synthesis
    pass
