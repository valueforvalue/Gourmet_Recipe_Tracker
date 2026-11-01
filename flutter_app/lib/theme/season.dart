// File: lib/theme/season.dart
//
// Maps the current month to a season (northern-hemisphere, astronomical).
// Mirrors the web UI's data-theme logic in web/src/Main.elm.

enum Season { spring, summer, autumn, winter }

/// Returns the season for a 1-based month (1 = January).
Season seasonFromMonth(int month) {
  if (month >= 3 && month <= 5) return Season.spring;
  if (month >= 6 && month <= 8) return Season.summer;
  if (month >= 9 && month <= 11) return Season.autumn;
  return Season.winter;
}

/// Season for the current calendar date.
Season currentSeason() => seasonFromMonth(DateTime.now().month);
