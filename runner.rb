require 'net/http'
require 'uri'
require 'json'

def open(url)
  Net::HTTP.get(URI.parse(url))
end

def send_to_telegram(message)
  apiToken = "1388326080:AAFGxulzcVRIJwSCcQr1pGjddyOwvC5_Fe0"
  chatID = '-1001675706309'
  apiURL = "https://api.telegram.org/bot#{apiToken}/sendMessage"

  # post to url
  uri = URI.parse(apiURL)
  https = Net::HTTP.new(uri.host, uri.port)
  https.use_ssl = true
  request = Net::HTTP::Post.new(uri.request_uri)
  request.set_form_data({"chat_id" => chatID, "text" => message})
  response = https.request(request)
end

school_id = "ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba"
# load tomorrow's date
date = Time.now + 86400

url = "https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId=ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba&ServingDate=12%2016%202022&ServingLine=Standard%20Line&MealType=Lunch"
page_content = open(url)
values = JSON.parse(page_content)

message = "Lunch for #{date.strftime("%A, %B %d, %Y")} is: \r\n\r\n"

lunch = values["LUNCH"]
lunch.each do |item|
  message += "#{item["MenuItemDescription"]}\r\n"
end
puts message
send_to_telegram(message)