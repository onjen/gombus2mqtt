# gombus2mqtt

[M-Bus](https://en.wikipedia.org/wiki/Meter-Bus) is a metering bus to read water, gas or electricity meters. To ingest the data into home automation systems like [home assistant](https://www.home-assistant.io/), gombus2mqtt publishes values from the M-Bus to MQTT.

# Getting Started

After configuring the M-Bus device in the `config.yaml` run a scan of the bus
```
gombus2mqtt-arm -s
```
This will scan all devices on the bus and print devices when they're found.
```
INFO Scanning for devices, this will take a while...
INFO Found device primary_address=3 serial_number=32008774 manufacturer=ZRI version=136 device_type="Heat: Outlet"
```
Afterwards print a full response for each device which was found, specified by the primary address.
```
gombus2mqtt-arm -p 3
```
Example response looks like so
```json
{
  "SerialNumber": 10000000,
  "Manufacturer": "ZRI",
  "ProductName": "",
  "Version": 136,
  "DeviceType": "Heat: Outlet",
  "AccessNumber": 0,
  "Signature": 0,
  "Status": 0,
  "DataRecords": [
    {
      "Function": "Instantaneous value",
      "StorageNumber": 0,
      "Tariff": 0,
      "Device": 0,
      "Unit": {
        "Exp": 1,
        "Unit": "none",
        "Type": 120,
        "VIFUnitDesc": ""
      },
      "Exponent": 0,
      "Type": "",
      "Quantity": "",
      "Value": 100000000,
      "ValueString": "",
      "RawValue": 10000000,
      "HasMoreRecords": false
    },
    {
      "Function": "Instantaneous value",
      "StorageNumber": 0,
      "Tariff": 0,
      "Device": 0,
      "Unit": {
        "Exp": 1000,
        "Unit": "WH",
        "Type": 7,
        "VIFUnitDesc": ""
      },
      "Exponent": 0,
      "Type": "",
      "Quantity": "",
      "Value": 11111,
      "ValueString": "",
      "RawValue": 11111,
      "HasMoreRecords": false
    },
    ...
  ]
}
```
You can see the values are part of the `DataRecords` list. Based on these values build up a meter template for your kind of meter, specifying a name and if the field should be published. Optionally you can also specify a `device_class` and a `unit` for Home Assistant autodiscovery.

# Home assistant MQTT Autodiscovery
By setting `homeassistant_autodiscover/enabled: true` in the config, `gombus2mqtt` will publish Home assistant MQTT integration compatible [Autodiscover](https://www.home-assistant.io/integrations/mqtt/#mqtt-discovery) messages. Make sure that the `device_class` and `unit` are set to [Home Assistant's Sensor standard](https://www.home-assistant.io/integrations/sensor/#device-class) compatible values.

The published values are then automatically created and populated as Sensors in Home Assistant.
