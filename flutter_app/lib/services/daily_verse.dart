// File: lib/services/daily_verse.dart
//
// A verse-of-the-day keyed to the calendar date (mirrors the web UI).

class DailyVerse {
  final String text;
  final String ref;
  const DailyVerse({required this.text, required this.ref});
}

const List<DailyVerse> _verses = [
  DailyVerse(
    text: 'Go, eat your bread with joy, and drink your wine with a merry heart, for God has already approved what you do.',
    ref: 'Ecclesiastes 9:7',
  ),
  DailyVerse(
    text: 'Whether you eat or drink, or whatever you do, do all to the glory of God.',
    ref: '1 Corinthians 10:31',
  ),
  DailyVerse(text: 'He fills your soul with good things.', ref: 'Psalm 103:5'),
  DailyVerse(
    text: 'Better is a dinner of herbs where love is than a fattened ox and hatred with it.',
    ref: 'Proverbs 15:17',
  ),
  DailyVerse(text: 'Give us this day our daily bread.', ref: 'Matthew 6:11'),
  DailyVerse(text: 'Taste and see that the Lord is good.', ref: 'Psalm 34:8'),
];

/// Today's verse, chosen by days-since-epoch modulo the verse count.
DailyVerse verseForToday([DateTime? now]) {
  final d = now ?? DateTime.now();
  final days = d.millisecondsSinceEpoch ~/ Duration.millisecondsPerDay;
  return _verses[days % _verses.length];
}
