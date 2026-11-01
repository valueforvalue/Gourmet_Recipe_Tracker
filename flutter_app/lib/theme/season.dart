// File: lib/theme/season.dart
//
// Maps the current month to a season (northern-hemisphere, astronomical).
// Mirrors the web UI's data-theme logic in web/src/Main.elm.

import 'package:flutter/foundation.dart';

enum Season {
  spring,
  summer,
  autumn,
  winter;

  String get label => name[0].toUpperCase() + name.substring(1);
}

/// Returns the season for a 1-based month (1 = January).
Season seasonFromMonth(int month) {
  if (month >= 3 && month <= 5) return Season.spring;
  if (month >= 6 && month <= 8) return Season.summer;
  if (month >= 9 && month <= 11) return Season.autumn;
  return Season.winter;
}

/// Season for the current calendar date.
Season currentSeason() => seasonFromMonth(DateTime.now().month);

/// The active season, changeable at runtime so Settings can preview themes.
final ValueNotifier<Season> activeSeason = ValueNotifier(currentSeason());
