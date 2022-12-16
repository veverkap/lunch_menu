require 'net/http'
require 'uri'
require 'json'
require 'date'
require 'time'
require 'logger'

LOGGER = Logger.new(STDOUT)

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
  LOGGER.info "Response: #{response.body}"
end

school_id = "ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba"
# load tomorrow's date
ENV['TZ'] = 'America/New_York'
date = Date.today + 1

LOGGER.info "Processing date #{date}"
if date.saturday? || date.sunday?
  send_to_telegram("Skipping weekend")
  LOGGER.info "Skipping weekend"
  exit
end

url = "https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId=ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba&ServingDate=12%2016%202022&ServingLine=Standard%20Line&MealType=Lunch"

LOGGER.info "Loading #{url}"

page_content = open(url)

LOGGER.info "Parsing JSON of #{page_content}"
values = JSON.parse(page_content)

message = "Lunch for #{date.strftime("%A, %B %d, %Y")} is: \r\n\r\n"

lunch = values["LUNCH"]
lunch.each do |item|
  message += "#{item["MenuItemDescription"]}\r\n"
end

LOGGER.info "Sending message: \r\n-----\r\n#{message}"

send_to_telegram(message)