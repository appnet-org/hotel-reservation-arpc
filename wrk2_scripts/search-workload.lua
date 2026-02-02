local socket = require("socket")
math.randomseed(socket.gettime()*1000)
math.random(); math.random(); math.random()

local url = "http://localhost:5000"

local function search_hotel() 
  local method = "GET"
  local path = url .. "/hotels?inDate=2015-04-10&outDate=2015-04-11&lat=38.0235&lon=-122.095"

  local headers = {}
  return wrk.format(method, path, headers, nil)
end

request = function()
  cur_time = math.floor(socket.gettime())
  return search_hotel(url)
end
