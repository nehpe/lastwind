# lastwind

A pair of CLI tools for checking local weather using the [National Weather Service API](https://www.weather.gov/documentation/services-web-api) — view current conditions and forecasts, or browse 3 days of observation history with wind extremes. Auto-detects your nearest station on first run.

## Installation

Requires [Go 1.21+](https://go.dev/dl/).

```sh
git clone https://github.com/nehpe/lastwind.git
cd lastwind
make build
```

This produces two binaries in the project root: `lastwind` and `forecast`.

## First Run

On first run, either command will auto-detect your location via IP geolocation, find your nearest NWS observation station, and save the configuration:

```
  ── lastwind configuration ──

  No configuration file found. Let's set one up.
  Detecting your location... found City, State
  Nearest station: ()

  ICAO station code []:
  Latitude []:
  Longitude []:

  Config saved to ~/.config/lastwind/config.json
```

Press Enter to accept the detected defaults, or type your own values. The config is stored at `~/.config/lastwind/config.json` and used for all subsequent runs.

## Commands

### `lastwind` — Observation History

Displays a table of recent weather observations and the highest wind/gust over the last 3 days.

```sh
./lastwind                    # use configured station
./lastwind -station KDEN      # override station
./lastwind -n 20              # show 20 most recent observations (default: 10)
```

```
  Station: Denver International Airport (KDEN)

  ┌────────────────┬────────────────┬────────┬──────┬──────┬────────┬──────────────────────────────┐
  │ Time           │ Wind           │ Vis mi │ Temp │ Dwpt │ Hum    │ Weather                      │
  ├────────────────┼────────────────┼────────┼──────┼──────┼────────┼──────────────────────────────┤
  │ Feb 17 10:53   │ W 20           │   10.0 │   52 │   23 │    32% │ Partly Cloudy                │
  │ Feb 17 09:53   │ W 20 G 25      │   10.0 │   47 │   29 │    50% │ Mostly Cloudy                │
  │ Feb 17 08:53   │ W 30 G 46      │    7.0 │   49 │   30 │    48% │ Mostly Cloudy and Windy      │
  └────────────────┴────────────────┴────────┴──────┴──────┴────────┴──────────────────────────────┘
  Showing 3 of 72 observations (3 days)

  ── 3-Day Extremes ─────────────────────────
  Highest Wind:  30 mph W (Feb 17 08:53)
  Highest Gust:  46 mph W (Feb 17 08:53)
```

### `forecast` — Current Conditions & Forecast

Shows current conditions at your nearest station and the forecast for today, tonight, tomorrow, and tomorrow night.

```sh
./forecast                              # use configured location
./forecast -lat 39.7392 -lon -104.9903  # override coordinates
```

```
  Glendale, CO
  Station: Denver International Airport (KDEN)

  ── Current Conditions (Feb 17 10:53) ──

    Partly Cloudy
    Temperature:  52°F
    Dewpoint:     23°F
    Humidity:     32%
    Wind:         W 20 mph
    Visibility:   10.0 mi
    Barometer:    29.49 in

  ── Forecast ───────────────────────────────

    Today              High: 55°F  Wind: SW 16 to 25 mph
      Mostly sunny. High near 55, with temperatures falling to
      around 49 in the afternoon. Southwest wind 16 to 25 mph,
      with gusts as high as 46 mph.

    Tonight            Low: 32°F  Wind: SW 8 to 18 mph
      Partly cloudy, with a low around 32. Southwest wind 8 to
      18 mph, with gusts as high as 30 mph.
```

## Configuration

The config file lives at `~/.config/lastwind/config.json`:

```json
{
  "station": "KDEN",
  "latitude": 39.8561,
  "longitude": -104.6737
}
```

Edit it directly or delete it to re-run the setup wizard. CLI flags (`-station`, `-lat`, `-lon`) always override the saved config.

## Development

```sh
make build       # build both binaries
make test        # run all tests (verbose)
make cover       # run tests with coverage summary
make cover-html  # generate HTML coverage report
make clean       # remove binaries and coverage files
```

## License

[GPLv3](LICENSE)
