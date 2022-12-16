import datetime
import requests
import urllib.parse

def send_to_telegram(message):
    apiToken = "1388326080:AAFGxulzcVRIJwSCcQr1pGjddyOwvC5_Fe0"
    chatID = '-1001675706309'
    apiURL = f'https://api.telegram.org/bot{apiToken}/sendMessage'

    try:
        response = requests.post(apiURL, json={'chat_id': chatID, 'text': message})
        print(response.text)
    except Exception as e:
        print(e)

# load from url
school_id = "ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba"
# load tomorrow's date
serving_date = (datetime.date.today() + datetime.timedelta(days=1))
# serving_date = datetime.date.today().strftime("%m/%d/%Y")
serve_date = serving_date.strftime("%m") + "%20" + serving_date.strftime("%d") + "%20" + serving_date.strftime("%Y")
print(serve_date)
url = f"https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId={school_id}&ServingDate={serve_date}&ServingLine=Standard%20Line&MealType=Lunch"
print(url)
# r = requests.get(url)
# data = r.json()
# print(data)

# message = f"The lunch menu for {serving_date} is:\r\n\r\n"

# for item in data["LUNCH"]:
#   message += item["MenuItemDescription"] + "\r\n"

# print(message)

# send_to_telegram(message)