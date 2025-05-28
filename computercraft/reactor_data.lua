local reactor = nil
local reactorVersion = nil

local function getPeripheral(typeName)
	for _, side in ipairs(peripheral.getNames()) do
		if peripheral.getType(side) == typeName then
			return side
		end
	end
	return nil
end

local function detectReactor()
	-- Bigger Reactors V1
	local reactor_bigger_v1 = getPeripheral("bigger-reactor")
	if reactor_bigger_v1 then
		reactor = peripheral.wrap(reactor_bigger_v1)
		reactorVersion = "Bigger Reactors"
		return true
	end

	-- Bigger Reactors V2
	local reactor_bigger_v2 = getPeripheral("BiggerReactors_Reactor")
	if reactor_bigger_v2 then
		reactor = peripheral.wrap(reactor_bigger_v2)
		reactorVersion = "Bigger Reactors"
		return true
	end

	-- Extreme or Big Reactors
	local reactor_extreme_or_big = getPeripheral("BigReactors-Reactor")
	if reactor_extreme_or_big then
		reactor = peripheral.wrap(reactor_extreme_or_big)
		reactorVersion = (reactor.mbIsConnected ~= nil) and "Extreme Reactors" or "Big Reactors"
		return true
	end

	return false
end

if not detectReactor() then
	print("No compatible reactor found.")
	return
end

print("Detected reactor: " .. reactorVersion)

local function getStats()
	local stats = {}

	stats.status = reactor.getActive and reactor.getActive() and "Running" or "Stopped"

	-- Core values
	stats.energyStored = reactor.getEnergyStored and reactor.getEnergyStored() or 0
	stats.energyProducedLastTick = reactor.getEnergyProducedLastTick and reactor.getEnergyProducedLastTick() or 0

	-- Temperature
	stats.fuelTemp = reactor.getFuelTemperature and reactor.getFuelTemperature() or 0
	stats.casingTemp = reactor.getCasingTemperature and reactor.getCasingTemperature() or 0

	-- Fuel & Waste
	stats.fuelAmount = reactor.getFuelAmount and reactor.getFuelAmount() or 0
	stats.wasteAmount = reactor.getWasteAmount and reactor.getWasteAmount() or 0
	stats.fuelConsumedLastTick = reactor.getFuelConsumedLastTick and reactor.getFuelConsumedLastTick() or 0
	stats.fuelReactivity = reactor.getFuelReactivity and reactor.getFuelReactivity() or 0
	stats.controlRodInsertion = reactor.getControlRodLevel and reactor.getControlRodLevel(0) or 0

	-- Optional label and ID
	stats.computerID = os.getComputerID()
	stats.computerLabel = os.getComputerLabel() or "unnamed"

	return stats
end

while true do
	local stats = getStats()
	local json = textutils.serializeJSON(stats)

	local res = http.post("http://localhost:8080/reactor", json, {
		["Content-Type"] = "application/json",
	})

	if res then
		print("Sent: " .. json)
		res.close()
	else
		print("Failed to send data")
	end

	sleep(5)
end
