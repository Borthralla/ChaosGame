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

type Flt_Point struct {
	X float64
	Y float64
}

//This function generates points using the chaos game algorithm
func make_points(num_points int, num_verts int, length int, iterations int) []image.Point {

	center := float64(length) / 2

	vertices := make_vertices(num_verts, length)

	points := make([]image.Point, num_points, num_points)

	for i := 0; i < num_points; i++ {
		points[i] = make_point(center, vertices, iterations)
	}

	return points
}





func make_vertices(num_verts int, length int) []Flt_Point {
	
	radius := float64(length) / 2


	vertices := make([]Flt_Point, num_verts, num_verts)
	
	var theta float64 = 2 * math.Pi / float64(num_verts)

	//Note that the rounding down is on purpose.
	

	for i := 0; i < num_verts; i++ {

		angle := float64(i) * theta

		dx := math.Cos(angle) * radius
		dy := math.Sin(angle) * radius

		//Note that the y axis is inverted
		vertex := Flt_Point{radius + dx, radius - dy}

		vertices[i] = vertex

	}

	return vertices
}

func make_point(center float64, vertices []Flt_Point, iterations int) image.Point {

	var vertex Flt_Point;

	current_x := center
	current_y := center

	for i := 0; i < iterations; i++ {
		vertex = vertices[rand.Intn(len(vertices))]
		current_x += (vertex.X - current_x) / 2
		current_y += (vertex.Y - current_y) / 2
	}

	return image.Point{int(current_x), int(current_y)}

}

func make_point_rand(center float64, vertices []Flt_Point, iterations int, r *rand.Rand) image.Point {

	var vertex Flt_Point;

	current_x := center
	current_y := center

	for i := 0; i < iterations; i++ {
		vertex = vertices[r.Intn(len(vertices))]
		current_x += (vertex.X - current_x) / 2
		current_y += (vertex.Y - current_y) / 2
	}

	return image.Point{int(current_x), int(current_y)}

}



func count_points(num_points int, num_verts int, length int, iterations int) ([]int, int)  {

	center := float64(length) / 2

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

func parallel_count_points(num_points int, num_verts int, length int, iterations int) ([]int, int) {

	center := float64(length) / 2

	vertices := make_vertices(num_verts, length)

	point_counts := make([]int, length * length, length * length)

	max_count := 0

	num_cpus := 4

	done := make(chan int, 4)

	for i := 0; i < num_cpus; i++ {
		go point_generator(num_points / num_cpus, point_counts, done, center, vertices, iterations, length, int64(i))
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

func point_generator(num_points int, point_counts []int, done chan int, center float64, vertices []Flt_Point, iterations int, length int, seed int64) {
	src := rand.NewSource(time.Now().UnixNano() + 2147483000 * seed)
	r := rand.New(src)
	for i := 0; i < num_points; i++ {
		point := make_point_rand(center, vertices, iterations, r)
		index := point.Y * length + point.X
		point_counts[index] += 1

		if (i & 33554431 == 0) {
			fmt.Printf("%f%% done\n", 100 * float64(i) / float64(num_points))
		}
	}

	done <- 1

}

func make_image(counts []int, max_count int, length int) {
	result := image.NewRGBA(image.Rect(0,0, length, length))

	max_count_flt := float64(max_count)

	for index, count := range counts {
		// Swapping x and y and negating y to make things stand vertically
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

func save_fractal(num_points int, num_sides int, length int, iterations int) {
	fmt.Printf("Making points\n")
	counts, max := parallel_count_points(num_points, num_sides, length, iterations)
	fmt.Printf("Making Image\n")
	make_image(counts, max, length)
}



func main() {

	rand.Seed(time.Now().UTC().UnixNano())
	save_fractal(10000000000, 9, 4000, 25)
}

// 6 seconds