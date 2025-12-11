#!/usr/bin/env python3

import json
import logging
import os
import sys
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Optional
from urllib.parse import quote

import pytz
import requests


@dataclass
class School:
    id: str
    name: str


# Configure JSON logging
class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_data = {
            "time": self.formatTime(record, self.datefmt),
            "level": record.levelname,
            "msg": record.getMessage(),
        }
        if hasattr(record, "extra_fields"):
            log_data.update(record.extra_fields)
        return json.dumps(log_data)


logger = logging.getLogger(__name__)
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(JSONFormatter())
logger.addHandler(handler)
logger.setLevel(logging.INFO)


telegram_token = ""
telegram_chesapeake_chat_id = ""
schools = [
    School(id="d9edb69f-dc06-41a4-8d8d-15c3e47d812f", name="Butts Road Intermediate"),
    School(id="6809b286-dbc7-48c1-bd22-d8db93816941", name="Butts Road Primary"),
]


def log_with_fields(level, msg, **kwargs):
    """Helper function to log with extra fields"""
    record = logger.makeRecord(
        logger.name,
        level,
        __file__,
        0,
        msg,
        (),
        None,
    )
    record.extra_fields = kwargs
    logger.handle(record)


def send_telegram_message(message: str) -> bool:
    """Send a message via Telegram"""
    url = f"https://api.telegram.org/bot{telegram_token}/sendMessage"
    payload = {
        "chat_id": telegram_chesapeake_chat_id,
        "text": message,
        "parse_mode": "Markdown",
    }

    try:
        resp = requests.post(url, json=payload)
        if resp.status_code != 200:
            log_with_fields(
                logging.ERROR,
                "Failed to send Telegram message",
                status=resp.status_code,
            )
            return False
        return True
    except Exception as e:
        log_with_fields(logging.ERROR, "Failed to send Telegram message", error=str(e))
        return False


def get_lunch_menu_for_school(date: datetime, school: School) -> Optional[str]:
    """Get lunch menu for a specific school"""
    log_with_fields(logging.INFO, "Getting lunch menu", date=date.isoformat(), school=school.name)

    url = (
        f"https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitemsByGrade"
        f"?SchoolId={school.id}"
        f"&ServingDate={quote(date.strftime('%m/%d/%Y'))}"
        f"&ServingLine=Main%20Line&MealType=Lunch&Grade=02&PersonId=null"
    )
    log_with_fields(logging.INFO, "Constructed URL", url=url)

    try:
        resp = requests.get(url)
        if resp.status_code != 200:
            log_with_fields(
                logging.ERROR,
                "Failed to get lunch menu",
                school=school.name,
                status=resp.status_code,
            )
            return None

        data = resp.json()
        menu_items = []

        if "ENTREES" in data:
            for item in data["ENTREES"]:
                description = item.get("MenuItemDescription", "")
                log_with_fields(logging.INFO, "Menu item", item=description)
                menu_items.append(description)

        if not menu_items:
            log_with_fields(logging.ERROR, "No menu items found", school=school.name)
            return None

        return "\n".join(menu_items) + "\n"

    except Exception as e:
        log_with_fields(
            logging.ERROR, "Failed to get lunch menu", school=school.name, error=str(e)
        )
        return None


def get_tomorrows_weather(date: datetime) -> Optional[str]:
    """Get weather forecast for tomorrow"""
    url = "https://wttr.in/Chesapeake?format=j1"
    log_with_fields(logging.INFO, "Getting weather", date=date.isoformat(), url=url)

    try:
        resp = requests.get(url)
        if resp.status_code != 200:
            log_with_fields(
                logging.ERROR, "Failed to get weather", status=resp.status_code
            )
            return None

        data = resp.json()
        weather_data = data.get("weather", [])

        for forecast in weather_data:
            forecast_date_str = forecast.get("date", "")
            try:
                forecast_date = datetime.strptime(forecast_date_str, "%Y-%m-%d")
                # Set to noon in America/New_York timezone to match Go behavior
                loc = pytz.timezone("America/New_York")
                forecast_date = loc.localize(
                    datetime(
                        forecast_date.year,
                        forecast_date.month,
                        forecast_date.day,
                        12,
                        0,
                        0,
                    )
                )

                if forecast_date.date() == date.date():
                    hourly_weather = []
                    hourly = forecast.get("hourly", [])

                    for hour in hourly:
                        time_str = hour.get("time", "")
                        if time_str in ["600", "1200", "1800"]:
                            temp_f = hour.get("tempF", "")
                            weather_desc_list = hour.get("weatherDesc", [])
                            weather_desc = (
                                weather_desc_list[0].get("value", "")
                                if weather_desc_list
                                else ""
                            )

                            time_label = {
                                "600": "6 AM",
                                "1200": "12 PM",
                                "1800": "6 PM",
                            }[time_str]

                            hourly_weather.append(
                                f"{time_label} - {temp_f}°F - {weather_desc}"
                            )

                    if hourly_weather:
                        return "\n".join(hourly_weather) + "\n"

            except ValueError:
                continue

        log_with_fields(logging.WARNING, "No weather found for date", date=date.isoformat())
        return None

    except Exception as e:
        log_with_fields(logging.ERROR, "Failed to get weather", error=str(e))
        return None


def main():
    global telegram_token, telegram_chesapeake_chat_id

    # Check for required environment variables
    telegram_chesapeake_chat_id = os.getenv("TELEGRAM_CHESAPEAKE_CHAT_ID", "")
    if not telegram_chesapeake_chat_id:
        log_with_fields(logging.ERROR, "TELEGRAM_CHESAPEAKE_CHAT_ID not set")
        return

    telegram_token = os.getenv("TELEGRAM_TOKEN", "")
    if not telegram_token:
        log_with_fields(logging.ERROR, "TELEGRAM_TOKEN not set")
        return

    # Get tomorrow's date in America/New_York timezone
    try:
        loc = pytz.timezone("America/New_York")
    except Exception as e:
        log_with_fields(
            logging.ERROR,
            "Failed to load location",
            location="America/New_York",
            error=str(e),
        )
        return

    # Get current time in NY timezone and set to noon to avoid DST issues
    now = datetime.now(loc)
    now = loc.localize(datetime(now.year, now.month, now.day, 12, 0, 0))
    tomorrow = now + timedelta(days=1)

    log_with_fields(
        logging.INFO,
        "Processing dates",
        now=now.isoformat(),
        tomorrow=tomorrow.isoformat(),
    )

    telegram_messages = []

    for school in schools:
        lunch_menu = get_lunch_menu_for_school(tomorrow, school)
        if lunch_menu:
            log_with_fields(
                logging.INFO, "Lunch menu", school=school.name, menu=lunch_menu
            )
            telegram_messages.append(f"*{school.name}*:\n{lunch_menu}")
        else:
            log_with_fields(logging.ERROR, "Failed to get lunch menu", school=school.name)

    if not telegram_messages:
        log_with_fields(logging.WARNING, "No lunch menus found for any school")
        return

    weather = get_tomorrows_weather(tomorrow)
    if not weather:
        log_with_fields(logging.ERROR, "Failed to get weather")
        weather = ""
    elif weather:
        telegram_messages.append(f"*Weather*:\n{weather}")

    # Build final message
    telegram_message = f"Lunch menu for {tomorrow.strftime('%m/%d/%Y')}:\n\n"
    telegram_message += "".join(telegram_messages)
    telegram_message += f"\n*Weather*:\n{weather}\n"

    if not send_telegram_message(telegram_message):
        log_with_fields(logging.ERROR, "Failed to send Telegram message")

    print(telegram_message)


if __name__ == "__main__":
    main()
