package tools

import (
	"math"
	"strconv"
	"ticketBookingSystem/models"
)

// type Seat struct {
// 	Type       string `json:"type"`
// 	Status     string `json:"status"`
// 	SeatNumber string `json:"seat_number"`
// }

// type SeatRow struct {
// 	Row   string `json:"row"`
// 	Seats []Seat `json:"seats"`
// }

// type Show struct {
// 	SeatRows []SeatRow `json:"seat_rows"`
// }

func isAvailable(seat models.Seat) bool {
	return seat.Type != "AISLE" && seat.Status == "available"
}

func createsIsolated(seats []models.Seat, start, k int) bool {
	end := start + k - 1

	// left side
	if start-1 >= 0 {
		if seats[start-1].Status == "available" {
			if start-2 < 0 || seats[start-2].Status != "available" {
				return true
			}
		}
	}

	// right side
	if end+1 < len(seats) {
		if seats[end+1].Status == "available" {
			if end+2 >= len(seats) || seats[end+2].Status != "available" {
				return true
			}
		}
	}

	return false
}

func FindBestSeats(show *models.Show, k int) ([]string, error) {
	rows := len(show.SeatRows)
	centerRow := rows / 2

	bestScore := math.MaxFloat64
	bestSeats := []string{}

	for r, row := range show.SeatRows {
		seats := row.Seats
		cols := len(seats)
		centerCol := cols / 2

		for c := 0; c <= cols-k; c++ {

			valid := true
			for i := 0; i < k; i++ {
				if !isAvailable(seats[c+i]) {
					valid = false
					break
				}
			}

			if !valid {
				continue
			}

			midCol := c + (k-1)/2

			score := math.Abs(float64(r-centerRow)) +
				math.Abs(float64(midCol-centerCol))

			if createsIsolated(seats, c, k) {
				score += 5
			}

			if score < bestScore {
				bestScore = score

				// 🔥 Build seat numbers
				currentSeats := []string{}
				seatIndex := 1

				for i := 0; i < len(seats); i++ {
					if seats[i].Type == "AISLE" {
						continue
					}

					if i >= c && i < c+k {
						seat := seats[i]

						// Use existing seat_number if present
						if seat.SeatNumber != "" {
							currentSeats = append(currentSeats, seat.SeatNumber)
						} else {
							// Generate like A1, A2...
							currentSeats = append(currentSeats,
								row.Row+strconv.Itoa(seatIndex))
						}
					}
					seatIndex++
				}

				bestSeats = currentSeats
			}
		}
	}

	return bestSeats, nil
}
