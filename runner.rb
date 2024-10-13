require 'cgi'
require 'date'
require 'json'
require 'logger'
require 'mailgun-ruby'
require 'net/http'
require 'stringio'
require 'time'
require 'uri'

TELEGRAM_CHESAPEAKE_CHAT_ID = ENV["TELEGRAM_CHESAPEAKE_CHAT_ID"]
SKIP_CHESAPEAKE             = ENV["SKIP_CHESAPEAKE"] == "true"

# loading logger
LOGGER = Logger.new(STDOUT)

def open(url)
  Net::HTTP.get(URI.parse(url))
end

def send_to_telegram(message, chat=TELEGRAM_CHESAPEAKE_CHAT_ID)
  LOGGER.info "Sending to Telegram"
  apiToken = ENV["TELEGRAM_TOKEN"]
  chatID   = chat
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

def send_to_email(email, message)
  # First, instantiate the Mailgun Client with your API key
  mg_client = Mailgun::Client.new(ENV['MAILGUN_API_KEY'])

  # Define your message parameters
  message_params =  {
    from: 'lunchmenu@mg.veverka.net',
    to:   email,
    subject: 'Lunch Menu',
    text:    message
  }

  # Send your message through the client
  mg_client.send_message "mg.veverka.net", message_params
end

def load_chesapeake_details(date, school_id="25128b5c-642c-461c-a224-3cc86a750b8f")
  ENV['TZ'] = 'America/New_York'
  url = StringIO.new
  url << "https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitemsByGrade?SchoolId=25128b5c-642c-461c-a224-3cc86a750b8f"#&ServingDate="
  url << "&ServingDate="
  url << CGI.escape(date.strftime("%m/%d/%Y"))#"12%2016%202022"
  url << "&ServingLine=Main%20Line&MealType=Lunch&Grade=02&PersonId=null"
  
  LOGGER.info "Loading #{url.string}"

  page_content = open(url.string)

  LOGGER.info "Parsing JSON"# of #{page_content}"
  values = JSON.parse(page_content)
  message = "Lunch for #{date.strftime("%A, %B %d, %Y")} is: \r\n\r\n"

  lunch = values["ENTREES"]
  lunch.each do |item|
    message += "#{item["MenuItemDescription"]}\r\n"
  end
  message
end

ENV['TZ'] = 'America/New_York'
date = Date.today + 1

LOGGER.info "Processing date #{date}"

if date.saturday? || date.sunday?
  LOGGER.info "Skipping weekend"
  exit
end

chesapeake = load_chesapeake_details(date, "25128b5c-642c-461c-a224-3cc86a750b8f")

LOGGER.info "Sending message for Chesapeake: \r\n-----\r\n#{chesapeake}"

if chesapeake.nil? || chesapeake.empty? || ENV["SKIP_CHESAPEAKE"] == 1
  LOGGER.info "No message for Chesapeake, skipping"
else
  send_to_telegram(chesapeake, TELEGRAM_CHESAPEAKE_CHAT_ID)

  email_addresses = ENV["EMAIL_ADDRESSES"].split(",")

  LOGGER.info "Sending to #{email_addresses.count} email addresses"

  email_addresses.each do |email|
    send_to_email(email, chesapeake)
  end
end
