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

nl = true
values["days"].each do |day|
  if day["date"] == search_date
    day["menu_items"].each do |item|
      if item["is_section_title"]
        message << "\r\n#{item["text"]}\r\n"
      end
      if item.key?("food")
        food = item["food"]
        if food.nil?
          if item.key?("text") && item["text"] == "with"
            message << " with "
            nl = false
          end
        else
          if food.key?("name")
            message << food["name"]
            message << "\r\n" unless nl
          end
        end
      end
    end
  end
end

LOGGER.info "Sending message: \r\n-----\r\n#{message}"

send_to_telegram(message)