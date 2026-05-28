@echo off
echo === LinguaForge ===

echo Starting API server...
start /B go run .\cmd\api\

echo Starting frontend...
cd frontend
start /B npm run dev
cd ..

echo.
echo Backend:  http://localhost:8080
echo Frontend: http://localhost:5173
echo.
echo Close the terminal to stop.
pause
