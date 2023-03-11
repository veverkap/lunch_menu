require 'cgi'
require 'date'
require 'json'
require 'logger'
require 'mailgun-ruby'
require 'net/http'
require 'stringio'
require 'time'
require 'uri'

TELEGRAM_SALEM_CHAT_ID = ENV["TELEGRAM_SALEM_CHAT_ID"]
TELEGRAM_CHAT_ID       = ENV["TELEGRAM_CHAT_ID"]
SKIP_CHESAPEAKE        = ENV["SKIP_CHESAPEAKE"] == "true"
SKIP_SALEM             = ENV["SKIP_SALEM"] == "true"

# loading logger
LOGGER = Logger.new(STDOUT)

def open(url)
  Net::HTTP.get(URI.parse(url))
end

def send_to_telegram(message, chat=TELEGRAM_CHAT_ID)
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

def load_chesapeake_details(date)
  url = StringIO.new

  url << "https://cpschools.api.nutrislice.com/menu/api/weeks/school/southeastern-elementary/menu-type/lunch/"
  url << date.strftime("%Y/%m/%d")
  url << "?format=json"

  LOGGER.info "Loading #{url.string}"

  page_content = open(url.string)

  LOGGER.info "Parsing JSON"# of #{page_content}"
  values = JSON.parse(page_content)

  message = "Lunch for #{date.strftime("%A, %B %d, %Y")} is: \r\n\r\n"

  search_date = date.strftime("%Y-%m-%d")

  menu_items = {}

  current_section = nil
  has_with = false

  LOGGER.info "Processing menu items"
  values["days"].each do |day|
    if day["date"] == search_date
      LOGGER.info "Found search date #{day["date"]}"
      day["menu_items"].each do |item|
        if item["is_section_title"]
          LOGGER.info "Found section #{item["text"]}"
          current_section = item["text"].gsub("Choose One", "").gsub("-", "").strip
          menu_items[current_section] = []
        else
          next if current_section.nil?

          if item.key?("food")
            food = item["food"]

            if food.nil?
              if item.key?("text") && item["text"] == "with"
                LOGGER.info "Found with"
                menu_items[current_section][-1] = menu_items[current_section].last + " with"
                has_with = true
              end
            else
              if food.key?("name")
                LOGGER.info "Found food name #{food["name"]}"
                if has_with
                  LOGGER.info "Adding to last item"
                  menu_items[current_section][-1] = menu_items[current_section].last + " " + food["name"]
                  has_with = false
                else
                  LOGGER.info "Adding to new item"
                  menu_items[current_section] << food["name"]
                end
              end
            end
          end
        end
      end
    end
  end

  LOGGER.info "Building message"
  menu_items.each do |section, items|
    message << "#{section}: \r\n"
    items.each do |item|
      message << "- #{item}\r\n"
    end
    message << "\r\n"
  end
  message
end

def load_salem_details(date)
  school_id = "4fc4596f-ac21-42b6-a9a5-4a24282ce4e7"
  ENV['TZ'] = 'America/New_York'
  url = StringIO.new

  url << "https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId="
  url << school_id
  url << "&ServingDate="
  url << CGI.escape(date.strftime("%m/%d/%Y"))#"12%2016%202022"
  url << "&ServingLine=Standard%20Line&MealType=Lunch"

  LOGGER.info "Loading #{url.string}"

  page_content = open(url.string)

  LOGGER.info "Parsing JSON"# of #{page_content}"
  values = JSON.parse(page_content)

  message = "Lunch for #{date.strftime("%A, %B %d, %Y")} is: \r\n\r\n"

  lunch = values["LUNCH"]
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

chesapeake = load_chesapeake_details(date)
salem = load_salem_details(date)

LOGGER.info "Sending message for Chesapeake: \r\n-----\r\n#{chesapeake}"
LOGGER.info "Sending message for Salem: \r\n-----\r\n#{salem}"

if chesapeake.nil? || chesapeake.empty? || ENV["SKIP_CHESAPEAKE"] == 1
  LOGGER.info "No message for Chesapeake, skipping"
else
  send_to_telegram(chesapeake, TELEGRAM_CHAT_ID)

  email_addresses = ENV["EMAIL_ADDRESSES"].split(",")

  LOGGER.info "Sending to #{email_addresses.count} email addresses"

  email_addresses.each do |email|
    send_to_email(email, chesapeake)
  end
end

if salem.nil? || salem.empty? || ENV["SKIP_SALEM"] == 1
  LOGGER.info "No message for Salem, skipping"
else
  send_to_telegram(salem, TELEGRAM_SALEM_CHAT_ID)
end
