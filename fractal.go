package main

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"image/color"
	"image/png"
	"time"
	"os"
)

//This function generates points using the chaos game algorithm
func make_points(num_points int, num_verts int, length int, iterations int) []image.Point {

	if (length % 2 == 0) {
		length -= 1
	}

	center := image.Point{length / 2, length / 2}

	vertices := make_vertices(num_verts, length)

	points := make([]image.Point, num_points, num_points)

	for i := 0; i < num_points; i++ {
		points[i] = make_point(center, vertices, iterations)
	}

	return points
}

func count_points(num_points int, num_verts int, length int, iterations int) (map[image.Point]int, int)  {
	if (length % 2 == 0) {
		length -= 1
	}

	center := image.Point{length / 2, length / 2}

	vertices := make_vertices(num_verts, length)

	point_counts := make(map[image.Point]int, length * length)

	max_count := 0

	for i := 0; i < num_points; i++ {
		point := make_point(center, vertices, iterations)

		count := point_counts[point]
		count += 1
		point_counts[point] = count
		if (count > max_count) {
			max_count = count
		}

	}

	return point_counts, max_count
}

func make_vertices(num_verts int, length int) []image.Point {
	if (length % 2 == 0) {
		length -= 1
	}

	center := image.Point{length / 2, length / 2}

	vertices := make([]image.Point, num_verts, num_verts)
	
	var theta float64 = 2 * math.Pi / float64(num_verts)

	//Note that the rounding down is on purpose.
	flt_length := float64(length / 2)

	for i := 0; i < num_verts; i++ {

		angle := float64(i) * theta

		dx := int(math.Round(math.Cos(angle) * flt_length))
		dy := int(math.Round(math.Sin(angle) * flt_length))

		//Note that the y axis is inverted
		vertex := image.Point{center.X + dx, center.Y - dy}

		vertices[i] = vertex

	}

	return vertices
}

func make_point(center image.Point, vertices []image.Point, iterations int) image.Point {

	var vertex image.Point;

	current_point := center

	for i := 0; i < iterations; i++ {
		vertex = vertices[rand.Intn(len(vertices))]
		current_point.X += int(math.Round(float64(vertex.X - current_point.X) / 2))
		current_point.Y += int(math.Round(float64(vertex.Y - current_point.Y) / 2))
	}

	return current_point

}

func make_point_rand(center image.Point, vertices []image.Point, iterations int, r *rand.Rand) image.Point {

	var vertex image.Point;

	current_point := center

	for i := 0; i < iterations; i++ {
		vertex = vertices[r.Intn(len(vertices))]
		current_point.X += int(math.Round(float64(vertex.X - current_point.X) / 2))
		current_point.Y += int(math.Round(float64(vertex.Y - current_point.Y) / 2))
	}

	return current_point

}

func parallel_count_points(num_points int, num_verts int, length int, iterations int) (map[image.Point]int, int) {
	all_counts := make(chan (map[image.Point]int), 4) // 4 cpus boi
	point_counts := make(map[image.Point]int, length * length)


	for i := 0; i < 4; i++ {
		go point_generator(num_points / 4, num_verts, length, iterations, all_counts)
	}

	max := 0

	for i := 0; i < 4; i++ {
		counts := <- all_counts
		fmt.Printf("%d\n", i)
		for point, count := range counts {
			new_count := count + point_counts[point]
			point_counts[point] = new_count
			if new_count > max {
				max = new_count
			}
		}
	}
	return point_counts, max


}

func arr_count_points(num_points int, num_verts int, length int, iterations int) ([]int, int)  {
	if (length % 2 == 0) {
		length -= 1
	}

	center := image.Point{length / 2, length / 2}

	vertices := make_vertices(num_verts, length)

	point_counts := make([]int, length * length, length * length)

	max_count := 0

	for i := 0; i < num_points; i++ {
		point := make_point(center, vertices, iterations)
		index := point.Y * length + point.X
		count := point_counts[index]
		count += 1
		point_counts[index] = count
		if (count > max_count) {
			max_count = count
		}
	}

	return point_counts, max_count
}

func parallel_arr_count_points(num_points int, num_verts int, length int, iterations int) ([]int, int) {
	if (length % 2 == 0) {
		length -= 1
	}

	center := image.Point{length / 2, length / 2}

	vertices := make_vertices(num_verts, length)

	point_counts := make([]int, length * length, length * length)

	max_count := 0

	num_cpus := 4

	done := make(chan int, 4)

	for i := 0; i < num_cpus; i++ {
		go arr_point_generator(num_points / num_cpus, point_counts, done, center, vertices, iterations, length)
	}

	for i := 0; i < num_cpus; i++ {
		<- done 
	}

	for i := 0; i < length * length; i++ {
		count := point_counts[i]
		if (count > max_count) {
			max_count = count
		}
	}

	return point_counts, max_count


}

func arr_point_generator(num_points int, point_counts []int, done chan int, center image.Point, vertices []image.Point, iterations int, length int) {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	for i := 0; i < num_points; i++ {
		point := make_point_rand(center, vertices, iterations, r)
		index := point.Y * length + point.X
		point_counts[index] += 1
	}

	done <- 1

}


func point_generator(num_points int, num_verts int, length int, iterations int, results chan map[image.Point]int) {
	counts, _ := count_points(num_points, num_verts, length, iterations)
	results <- counts
}

func make_image(counts map[image.Point]int, max_count int, length int) {
	result := image.NewRGBA(image.Rect(0,0, length, length))

	max_count_flt := float64(max_count)

	for point, count := range counts {
		p := uint8(math.Round(255 * (1 - float64(count) / max_count_flt)))
		result.Set(point.X, point.Y, color.RGBA{p,p,p,255})
	}

	fmt.Printf("Writing image")
	outputFile, err := os.Create("test.png")

	if err != nil {
    	fmt.Printf("%s", err)
    	return
    }

    png.Encode(outputFile, result)

}

func arr_make_image(counts []int, max_count int, length int) {
	result := image.NewRGBA(image.Rect(0,0, length, length))

	if length % 2 == 0 {
		length -= 1
	}

	max_count_flt := float64(max_count)

	for index, count := range counts {
		x := index / length
		y := length - 1 - (index % length)
		p := uint8(math.Round(255 * (1 - float64(count) / max_count_flt)))
		result.Set(x, y, color.RGBA{p,p,p,255})
	}

	fmt.Printf("Writing image")
	outputFile, err := os.Create("test.png")

	if err != nil {
    	fmt.Printf("%s", err)
    	return
    }

    png.Encode(outputFile, result)
}



func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Printf("Making points\n")
	counts, max := parallel_arr_count_points(300000000, 5, 2000, 25)
	fmt.Printf("Making Image\n")
	arr_make_image(counts, max, 2000)
}

// 6 seconds