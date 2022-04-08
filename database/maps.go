package database

var (
	DKMaps = map[int16][]int16{
		18: {18, 193, 200}, 19: {19, 194, 201}, 25: {25, 195, 202}, 26: {26, 196, 203}, 27: {27, 197, 204}, 29: {29, 198, 205}, 30: {30, 199, 206}, // Normal Maps
		193: {18, 193, 200}, 194: {19, 194, 201}, 195: {25, 195, 202}, 196: {26, 196, 203}, 197: {27, 197, 204}, 198: {29, 198, 205}, 199: {30, 199, 206}, // DK Maps
		200: {18, 193, 200}, 201: {19, 194, 201}, 202: {25, 195, 202}, 203: {26, 196, 203}, 204: {27, 197, 204}, 205: {29, 198, 205}, 206: {30, 199, 206}, // Normal Maps
	}

	sharedMaps = []int16{1, 2, 3, 14, 15, 20, 21, 22, 23, 24, 26, 27, 32, 33, 34, 36, 37, 38, 42, 43, 44, 45, 46, 47,
		89, 93, 94, 95, 100, 101, 102, 108, 109, 110, 111, 112, 121, 122, 123, 160, 161, 162, 163, 164, 165, 166, 167, 168, 169, 170,
		213, 214, 221, 222, 223, 224, 225, 226, 227, 228, 233, 236, 237, 238, 239, 240, 241, 244, 252, 254, 251, 210}

	DungeonZones = []int16{229}

	PvPZones     = []int16{0, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159, 160, 161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 187, 188, 189, 190, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255}
	unlockedMaps = []int16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159, 160, 161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 187, 188, 189, 190, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255}

	ZhuangFactionMobs = []int{424203, 424204, 424205, 424206, 424207, 41766, 424201, //great war mobs
		425101, 425102, 425103, 425104, 425105, 425106, 425107, 425108, 425109, 425501, 425502, 425503, 425504} //faction war mobs
	ShaoFactionMobs = []int{424203, 424204, 424205, 424206, 424207, 41767, 424202,
		425110, 425111, 425112, 425113, 425114, 425115, 425116, 425117, 425118, 425505, 425506, 425507, 425508} //great war mobs
)