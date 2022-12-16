import datetime
import requests

# load from url
school_id = "ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba"
# load tomorrow's date
serving_date = (datetime.date.today() + datetime.timedelta(days=1)).strftime("%m/%d/%Y")
# serving_date = datetime.date.today().strftime("%m/%d/%Y")
url = f"https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId={school_id}&ServingDate={serving_date}&ServingLine=Standard%20Line&MealType=Lunch"
r = requests.get(url)
data = r.json()

message = f"The lunch menu for {serving_date} is:\r\n\r\n"

for item in data["LUNCH"]:
  message += item["MenuItemDescription"] + "\r\n"

print(message)


def send_to_telegram(message):
    apiToken = "1388326080:AAFGxulzcVRIJwSCcQr1pGjddyOwvC5_Fe0"
    chatID = '-1001675706309'
    apiURL = f'https://api.telegram.org/bot{apiToken}/sendMessage'

    try:
        response = requests.post(apiURL, json={'chat_id': chatID, 'text': message})
        print(response.text)
    except Exception as e:
        print(e)

send_to_telegram(message)