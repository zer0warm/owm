# owm

This program gets current weather data. It uses OpenWeathermap's [current weather data API](https://openweathermap.org/current). The original purpose is to use with [i3status](https://i3wm.org/docs/i3status.html), so the result is very short. The format is: `[Main weather group]: Actual_temp Feels_like_temp`.

It reads `$HOME/.curwttr_env` for API sensitive parameters. The format is:
```
appid=
lat=
lon=
```

The program is scheduled to run with `cron`. Successful execution puts the result in `$HOME/.curwttr`. If there are errors, they are logged to `$HOME/.curwttr_error`. From here my i3status configuration reads from `$HOME/.curwttr` and displays on i3bar.
