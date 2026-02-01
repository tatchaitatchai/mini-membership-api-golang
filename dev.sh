#!/bin/bash
# Kill any process on port 8085 before starting air
lsof -ti:8085 | xargs kill -9 2>/dev/null || true
sleep 0.5
air
