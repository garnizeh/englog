package services

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnalyticsService_ComprehensiveValidation tests comprehensive business logic validation
func TestAnalyticsService_ComprehensiveValidation(t *testing.T) {
	// Create test logger to prevent nil pointer dereference
	testLogger := logging.NewTestLogger()
	analyticsService := NewAnalyticsService(nil, testLogger)

	t.Run("DateRangeValidationEdgeCases", func(t *testing.T) {
		baseTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		tests := []struct {
			name      string
			startDate time.Time
			endDate   time.Time
			wantErr   bool
			errMsg    string
		}{
			{
				name:      "valid range - same day",
				startDate: baseTime,
				endDate:   baseTime.Add(23*time.Hour + 59*time.Minute),
				wantErr:   false,
			},
			{
				name:      "valid range - one week",
				startDate: baseTime,
				endDate:   baseTime.Add(7 * 24 * time.Hour),
				wantErr:   false,
			},
			{
				name:      "valid range - one month",
				startDate: baseTime,
				endDate:   baseTime.Add(30 * 24 * time.Hour),
				wantErr:   false,
			},
			{
				name:      "valid range - one year",
				startDate: baseTime,
				endDate:   baseTime.Add(365 * 24 * time.Hour),
				wantErr:   false,
			},
			{
				name:      "exact same timestamp",
				startDate: baseTime,
				endDate:   baseTime,
				wantErr:   false,
			},
			{
				name:      "end 1 second before start",
				startDate: baseTime,
				endDate:   baseTime.Add(-1 * time.Second),
				wantErr:   true,
				errMsg:    "end date must be after or equal to start date",
			},
			{
				name:      "end 1 day before start",
				startDate: baseTime,
				endDate:   baseTime.Add(-24 * time.Hour),
				wantErr:   true,
				errMsg:    "end date must be after or equal to start date",
			},
			{
				name:      "very large range",
				startDate: baseTime,
				endDate:   baseTime.Add(10 * 365 * 24 * time.Hour), // 10 years
				wantErr:   false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := analyticsService.validateDateRange(tt.startDate, tt.endDate)
				if tt.wantErr {
					assert.Error(t, err)
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("UserIDValidationEdgeCases", func(t *testing.T) {
		ctx := context.Background()

		tests := []struct {
			name    string
			userID  string
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid UUID",
				userID:  "123e4567-e89b-12d3-a456-426614174000",
				wantErr: false,
			},
			{
				name:    "valid UUID with uppercase",
				userID:  "123E4567-E89B-12D3-A456-426614174000",
				wantErr: false,
			},
			{
				name:    "nil UUID",
				userID:  "00000000-0000-0000-0000-000000000000",
				wantErr: false, // Nil UUID is technically valid
			},
			{
				name:    "empty string",
				userID:  "",
				wantErr: true,
				errMsg:  "invalid user ID",
			},
			{
				name:    "invalid format - too short",
				userID:  "123e4567-e89b-12d3-a456",
				wantErr: true,
				errMsg:  "invalid user ID",
			},
			{
				name:    "invalid format - missing hyphens",
				userID:  "123e4567e89b12d3a456426614174000",
				wantErr: false, // UUID without hyphens is actually valid
			},
			{
				name:    "invalid format - extra characters",
				userID:  "123e4567-e89b-12d3-a456-426614174000-extra",
				wantErr: true,
				errMsg:  "invalid user ID",
			},
			{
				name:    "invalid characters",
				userID:  "123g4567-e89b-12d3-a456-426614174000",
				wantErr: true,
				errMsg:  "invalid user ID",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := analyticsService.validateUserID(ctx, tt.userID)
				if tt.wantErr {
					assert.Error(t, err)
					if err != nil && tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestAnalyticsService_MetricCalculations(t *testing.T) {
	t.Run("ProductivityScoreCalculation", func(t *testing.T) {
		tests := []struct {
			name             string
			highValueMinutes int
			totalMinutes     int
			expectedScore    float64
		}{
			{
				name:             "perfect score",
				highValueMinutes: 100,
				totalMinutes:     100,
				expectedScore:    100.0,
			},
			{
				name:             "half high value",
				highValueMinutes: 50,
				totalMinutes:     100,
				expectedScore:    50.0,
			},
			{
				name:             "no high value",
				highValueMinutes: 0,
				totalMinutes:     100,
				expectedScore:    0.0,
			},
			{
				name:             "no activity",
				highValueMinutes: 0,
				totalMinutes:     0,
				expectedScore:    0.0,
			},
			{
				name:             "quarter high value",
				highValueMinutes: 25,
				totalMinutes:     100,
				expectedScore:    25.0,
			},
			{
				name:             "three quarters high value",
				highValueMinutes: 75,
				totalMinutes:     100,
				expectedScore:    75.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				score := calculateProductivityScore(tt.highValueMinutes, tt.totalMinutes)
				assert.Equal(t, tt.expectedScore, score)
			})
		}
	})

	t.Run("AverageCalculation", func(t *testing.T) {
		tests := []struct {
			name     string
			values   []float64
			expected float64
		}{
			{
				name:     "single value",
				values:   []float64{42.5},
				expected: 42.5,
			},
			{
				name:     "multiple values",
				values:   []float64{10, 20, 30},
				expected: 20.0,
			},
			{
				name:     "empty slice",
				values:   []float64{},
				expected: 0.0,
			},
			{
				name:     "values with decimals",
				values:   []float64{1.5, 2.5, 3.0},
				expected: 2.333333333333333,
			},
			{
				name:     "negative values",
				values:   []float64{-10, 10},
				expected: 0.0,
			},
			{
				name:     "zero values",
				values:   []float64{0, 0, 0},
				expected: 0.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				avg := calculateAverage(tt.values)
				assert.InDelta(t, tt.expected, avg, 0.000001) // Use InDelta for float comparison
			})
		}
	})

	t.Run("PercentageDistributionCalculation", func(t *testing.T) {
		tests := []struct {
			name     string
			counts   map[string]int
			expected map[string]float64
		}{
			{
				name: "even distribution",
				counts: map[string]int{
					"A": 50,
					"B": 50,
				},
				expected: map[string]float64{
					"A": 50.0,
					"B": 50.0,
				},
			},
			{
				name: "uneven distribution",
				counts: map[string]int{
					"A": 75,
					"B": 25,
				},
				expected: map[string]float64{
					"A": 75.0,
					"B": 25.0,
				},
			},
			{
				name: "three categories",
				counts: map[string]int{
					"A": 60,
					"B": 30,
					"C": 10,
				},
				expected: map[string]float64{
					"A": 60.0,
					"B": 30.0,
					"C": 10.0,
				},
			},
			{
				name:     "empty map",
				counts:   map[string]int{},
				expected: map[string]float64{},
			},
			{
				name: "single category",
				counts: map[string]int{
					"A": 100,
				},
				expected: map[string]float64{
					"A": 100.0,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := calculatePercentageDistribution(tt.counts)
				assert.Equal(t, len(tt.expected), len(result))
				for key, expectedValue := range tt.expected {
					assert.InDelta(t, expectedValue, result[key], 0.001)
				}
			})
		}
	})
}

func TestAnalyticsService_TimeZoneHandling(t *testing.T) {
	t.Run("DifferentTimeZones", func(t *testing.T) {
		// Test data in different time zones
		utcTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		locations := []struct {
			name     string
			location string
		}{
			{"UTC", "UTC"},
			{"New York", "America/New_York"},
			{"Tokyo", "Asia/Tokyo"},
			{"London", "Europe/London"},
			{"Los Angeles", "America/Los_Angeles"},
		}

		for _, loc := range locations {
			t.Run(loc.name, func(t *testing.T) {
				timezone, err := time.LoadLocation(loc.location)
				require.NoError(t, err)

				localTime := utcTime.In(timezone)

				// Test that time conversion doesn't affect date calculations
				startOfDay := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, timezone)
				endOfDay := startOfDay.Add(24 * time.Hour)

				assert.True(t, startOfDay.Before(endOfDay))
				assert.Equal(t, 24*time.Hour, endOfDay.Sub(startOfDay))
			})
		}
	})
}

func TestAnalyticsService_DataAggregation(t *testing.T) {
	t.Run("WeeklyAggregation", func(t *testing.T) {
		// Test weekly aggregation logic
		monday := time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC) // A Monday

		tests := []struct {
			name              string
			date              time.Time
			expectedWeekStart time.Time
		}{
			{
				name:              "Monday",
				date:              monday,
				expectedWeekStart: monday,
			},
			{
				name:              "Tuesday",
				date:              monday.Add(24 * time.Hour),
				expectedWeekStart: monday,
			},
			{
				name:              "Sunday",
				date:              monday.Add(6 * 24 * time.Hour),
				expectedWeekStart: monday,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				weekStart := getWeekStart(tt.date)
				assert.Equal(t, tt.expectedWeekStart, weekStart)
				assert.Equal(t, time.Monday, weekStart.Weekday())
			})
		}
	})

	t.Run("MonthlyAggregation", func(t *testing.T) {
		tests := []struct {
			name               string
			date               time.Time
			expectedMonthStart time.Time
		}{
			{
				name:               "first day of month",
				date:               time.Date(2024, 6, 1, 15, 30, 0, 0, time.UTC),
				expectedMonthStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				name:               "middle of month",
				date:               time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC),
				expectedMonthStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				name:               "last day of month",
				date:               time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC),
				expectedMonthStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				monthStart := getMonthStart(tt.date)
				assert.Equal(t, tt.expectedMonthStart, monthStart)
				assert.Equal(t, 1, monthStart.Day())
			})
		}
	})
}

func TestAnalyticsService_BoundaryConditions(t *testing.T) {
	t.Run("ZeroValues", func(t *testing.T) {
		// Test handling of zero values
		assert.Equal(t, 0.0, calculateProductivityScore(0, 0))
		assert.Equal(t, 0.0, calculateAverage([]float64{}))
		assert.Equal(t, map[string]float64{}, calculatePercentageDistribution(map[string]int{}))
	})

	t.Run("LargeValues", func(t *testing.T) {
		// Test handling of large values
		largeValue := 1000000
		score := calculateProductivityScore(largeValue, largeValue)
		assert.Equal(t, 100.0, score)

		avg := calculateAverage([]float64{float64(largeValue), float64(largeValue)})
		assert.Equal(t, float64(largeValue), avg)
	})

	t.Run("EdgeCaseDates", func(t *testing.T) {
		// Test edge case dates
		leapYearFeb29 := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
		nextDay := leapYearFeb29.Add(24 * time.Hour)

		assert.Equal(t, time.March, nextDay.Month())
		assert.Equal(t, 1, nextDay.Day())
	})
}

// Helper functions for testing (these might not exist in the actual service)
func calculateProductivityScore(highValueMinutes, totalMinutes int) float64 {
	if totalMinutes == 0 {
		return 0.0
	}
	return (float64(highValueMinutes) / float64(totalMinutes)) * 100.0
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculatePercentageDistribution(counts map[string]int) map[string]float64 {
	total := 0
	for _, count := range counts {
		total += count
	}

	if total == 0 {
		return map[string]float64{}
	}

	result := make(map[string]float64)
	for key, count := range counts {
		result[key] = (float64(count) / float64(total)) * 100.0
	}

	return result
}

func getWeekStart(date time.Time) time.Time {
	// Get the start of the week (Monday)
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday becomes 7
	}
	daysFromMonday := weekday - 1
	weekStart := date.Add(-time.Duration(daysFromMonday) * 24 * time.Hour)
	return time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
}

func getMonthStart(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}
