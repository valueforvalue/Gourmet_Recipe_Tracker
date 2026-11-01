// File: test/season_test.dart
//
// Table-driven test for the month → season mapping.

import 'package:flutter_test/flutter_test.dart';
import 'package:morris_recipe_tracker/theme/season.dart';

void main() {
  test('seasonFromMonth maps astronomical seasons (northern hemisphere)', () {
    final cases = <int, Season>{
      1: Season.winter,
      2: Season.winter,
      3: Season.spring,
      4: Season.spring,
      5: Season.spring,
      6: Season.summer,
      7: Season.summer,
      8: Season.summer,
      9: Season.autumn,
      10: Season.autumn,
      11: Season.autumn,
      12: Season.winter,
    };

    for (final entry in cases.entries) {
      expect(seasonFromMonth(entry.key), entry.value,
          reason: 'month ${entry.key}');
    }
  });
}
