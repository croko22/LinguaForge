#!/usr/bin/env bash
set -e

echo "=== LinguaForge ==="

# Start backend
echo "Starting API server..."
go run ./cmd/api/ &
API_PID=$!

# Start frontend
echo "Starting frontend..."
cd frontend && npm run dev &
UI_PID=$!

# Cleanup on exit
trap "kill $API_PID $UI_PID 2>/dev/null" EXIT

echo ""
echo "Backend:  http://localhost:8080"
echo "Frontend: http://localhost:5173"
echo ""
echo "Press Ctrl+C to stop"

wait
