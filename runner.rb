require 'net/http'
require 'uri'

def open(url)
  Net::HTTP.get(URI.parse(url))
end

school_id = "ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba"
# load tomorrow's date
date = Time.now + 86400

url = "https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId=ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba&ServingDate=12%2016%202022&ServingLine=Standard%20Line&MealType=Lunch"
page_content = open(url)
puts page_content