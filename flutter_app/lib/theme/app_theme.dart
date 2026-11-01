// File: lib/theme/app_theme.dart
//
// Seasonal Material themes. Color values mirror the web SPA's CSS variables
// (web/css/app.css) so both clients feel like the same product.

import 'package:flutter/material.dart';
import 'season.dart';

class SeasonColors {
  final Color primary;
  final Color secondary;
  final Color accent;
  final Color background;

  const SeasonColors({
    required this.primary,
    required this.secondary,
    required this.accent,
    required this.background,
  });
}

const Map<Season, SeasonColors> _seasonColors = {
  Season.spring: SeasonColors(
    primary: Color(0xFF3d7a52),
    secondary: Color(0xFF7f9d5a),
    accent: Color(0xFFcf6f9b),
    background: Color(0xFFf2f7ed),
  ),
  Season.summer: SeasonColors(
    primary: Color(0xFF14689e),
    secondary: Color(0xFF2e8fc4),
    accent: Color(0xFFf0652f),
    background: Color(0xFFeff7fd),
  ),
  Season.autumn: SeasonColors(
    primary: Color(0xFF9c4a1f),
    secondary: Color(0xFFb0763b),
    accent: Color(0xFFd97a38),
    background: Color(0xFFfdf3ea),
  ),
  Season.winter: SeasonColors(
    primary: Color(0xFF2c4a6e),
    secondary: Color(0xFF5a7d9e),
    accent: Color(0xFF7f9db8),
    background: Color(0xFFf2f6fb),
  ),
};

/// Colors for a season.
SeasonColors colorsFor(Season season) => _seasonColors[season]!;

/// Builds the Material theme for [season].
ThemeData themeFor(Season season) {
  final c = colorsFor(season);
  final scheme = ColorScheme.fromSeed(seedColor: c.primary).copyWith(
    primary: c.primary,
    secondary: c.secondary,
    tertiary: c.accent,
    surface: c.background,
  );
  return ThemeData(
    colorScheme: scheme,
    useMaterial3: true,
    scaffoldBackgroundColor: c.background,
  );
}
