if os.getenv("CI") then
  -- travis
else
  --dofile("./test/test_cloudwatch_logs.lua")
  dofile("./test/test_cloudwatch_get_metric_data.lua")
end
