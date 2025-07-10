#!/usr/bin/env python3
import os
import sys
import subprocess
import time

def run_command_simple(cmd, timeout=60):
    """Простое выполнение команды с таймаутом"""
    try:
        print(f"Выполняю: {cmd}")
        result = subprocess.run(cmd, shell=True, timeout=timeout)
        if result.returncode == 0:
            print("✅ Успешно")
            return True
        else:
            print(f"❌ Ошибка (код: {result.returncode})")
            return False
    except subprocess.TimeoutExpired:
        print(f"⏰ Команда зависла: {cmd}")
        return False
    except Exception as e:
        print(f"❌ Исключение: {e}")
        return False

def main():
    print("=" * 50)
    print("   🚀 ПРОСТОЙ ЗАПУСК СИСТЕМЫ")
    print("=" * 50)
    print()

    # Простая проверка Docker
    print("Проверяем Docker...")
    if not run_command_simple("docker --version", 10):
        print("Docker не работает!")
        input("Нажмите Enter для выхода...")
        sys.exit(1)

    print()
    print("🚀 Пытаемся запустить систему...")
    print("Это может занять несколько минут...")
    
    # Пробуем сразу запустить без всяких проверок
    if run_command_simple("docker-compose up -d --build", 600):  # 10 минут на сборку
        print()
        print("=" * 50)
        print("   ✅ ВРОДЕ ЗАПУСТИЛОСЬ!")
        print("=" * 50)
        print()
        print("Проверь в браузере:")
        print("   🌐 http://localhost:8080")
        print()
        print("Если не работает, смотри логи:")
        print("   docker-compose logs")
    else:
        print()
        print("❌ Что-то пошло не так...")
        print("Попробуй вручную:")
        print("   docker-compose up -d --build")
    
    print()
    input("Нажмите Enter для выхода...")

if __name__ == "__main__":
    main()
