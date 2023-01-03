require 'net/http'
require 'uri'
require 'json'
require 'date'
require 'time'
require 'logger'
require 'stringio'
require 'cgi'

LOGGER = Logger.new(STDOUT)

def open(url)
  Net::HTTP.get(URI.parse(url))
end

def send_to_telegram(message)
  apiToken = ENV["TELEGRAM_TOKEN"]
  chatID   = ENV["TELEGRAM_CHAT_ID"]
  apiURL   = "https://api.telegram.org/bot#{apiToken}/sendMessage"

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
  LOGGER.info "Skipping weekend"
  exit
end


url = StringIO.new

url << "https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId="
url << school_id
url << "&ServingDate="
url << CGI.escape(date.strftime("%m/%d/%Y"))#"12%2016%202022"
url << "&ServingLine=Standard%20Line&MealType=Lunch"

LOGGER.info "Loading #{url.string}"

page_content = open(url.string)

LOGGER.info "Parsing JSON of #{page_content}"
values = JSON.parse(page_content)

message = "Lunch for #{date.strftime("%A, %B %d, %Y")} is: \r\n\r\n"

lunch = values["LUNCH"]
lunch.each do |item|
  message += "#{item["MenuItemDescription"]}\r\n"
end

LOGGER.info "Sending message: \r\n-----\r\n#{message}"

send_to_telegram(message)