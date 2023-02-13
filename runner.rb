require 'cgi'
require 'date'
require 'json'
require 'logger'
require 'mailgun-ruby'
require 'net/http'
require 'stringio'
require 'time'
require 'uri'

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

ENV['TZ'] = 'America/New_York'
date = Date.today

LOGGER.info "Processing date #{date}"

if date.saturday? || date.sunday?
  LOGGER.info "Skipping weekend"
  exit
end

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

values["days"].each do |day|
  if day["date"] == search_date
    day["menu_items"].each do |item|
      if item["is_section_title"]
        current_section = item["text"].gsub("Choose One", "").gsub("-", "").strip
        menu_items[current_section] = []
      else
        next if current_section.nil?

        if item.key?("food")
          food = item["food"]

          if food.nil?
            if item.key?("text") && item["text"] == "with"
              menu_items[current_section][-1] = menu_items[current_section].last + " with"
              has_with = true
            end
          else
            if food.key?("name")
              if has_with
                menu_items[current_section][-1] = menu_items[current_section].last + " " + food["name"]
                has_with = false
              else
                menu_items[current_section] << food["name"]
              end
            end
          end
        end
      end
    end
  end
end
menu_items.each do |section, items|
  message << "#{section}: \r\n"
  items.each do |item|
    message << "- #{item}\r\n"
  end
  message << "\r\n"
end

LOGGER.info "Sending message: \r\n-----\r\n#{message}"

send_to_telegram(message)

puts "Sending to emails: #{ENV["EMAIL_ADDRESSES"]}"
email_addresses = ENV["EMAIL_ADDRESSES"].split(",")

email_addresses.each do |email|
  send_to_email(email, message)
end