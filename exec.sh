#!/bin/bash

while true; do
  echo "Starting the program..."
  ./online-lecture-bot  # 실제 프로그램 실행 파일 이름으로 변경하세요.

  # 프로그램이 종료될 때까지 기다립니다.
  program_exit_code=$?

  echo "Program exited with code: $program_exit_code"

  # 비정상 종료(panic)인 경우 재시작합니다.  원하는 종료 코드를 추가하거나 제거하여 조건을 조정할 수 있습니다.
  if [ $program_exit_code -ne 0 ]; then
    echo "Restarting the program in 5 seconds..."
    sleep 5
  else
    echo "Program exited normally. Exiting the script."
    exit 0
  fi
done