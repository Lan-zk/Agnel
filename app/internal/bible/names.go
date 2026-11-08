package bible

// The canon is fixed at 66 books, so the Chinese names and the short glyphs the
// interface shows live here instead of in the database (which carries the
// English names the translations use).

var chinese = map[string]string{
	"Gen": "创世记", "Exod": "出埃及记", "Lev": "利未记", "Num": "民数记",
	"Deut": "申命记", "Josh": "约书亚记", "Judg": "士师记", "Ruth": "路得记",
	"1Sam": "撒母耳记上", "2Sam": "撒母耳记下", "1Kgs": "列王纪上", "2Kgs": "列王纪下",
	"1Chr": "历代志上", "2Chr": "历代志下", "Ezra": "以斯拉记", "Neh": "尼希米记",
	"Esth": "以斯帖记", "Job": "约伯记", "Ps": "诗篇", "Prov": "箴言",
	"Eccl": "传道书", "Song": "雅歌", "Isa": "以赛亚书", "Jer": "耶利米书",
	"Lam": "耶利米哀歌", "Ezek": "以西结书", "Dan": "但以理书", "Hos": "何西阿书",
	"Joel": "约珥书", "Amos": "阿摩司书", "Obad": "俄巴底亚书", "Jonah": "约拿书",
	"Mic": "弥迦书", "Nah": "那鸿书", "Hab": "哈巴谷书", "Zeph": "西番雅书",
	"Hag": "哈该书", "Zech": "撒迦利亚书", "Mal": "玛拉基书",
	"Matt": "马太福音", "Mark": "马可福音", "Luke": "路加福音", "John": "约翰福音",
	"Acts": "使徒行传", "Rom": "罗马书", "1Cor": "哥林多前书", "2Cor": "哥林多后书",
	"Gal": "加拉太书", "Eph": "以弗所书", "Phil": "腓立比书", "Col": "歌罗西书",
	"1Thess": "帖撒罗尼迦前书", "2Thess": "帖撒罗尼迦后书", "1Tim": "提摩太前书",
	"2Tim": "提摩太后书", "Titus": "提多书", "Phlm": "腓利门书", "Heb": "希伯来书",
	"Jas": "雅各书", "1Pet": "彼得前书", "2Pet": "彼得后书", "1John": "约翰一书",
	"2John": "约翰二书", "3John": "约翰三书", "Jude": "犹大书", "Rev": "启示录",
}

// Glyph is the one- or two-character mark on a book tile. Books whose Chinese
// name starts alike (撒母耳记上/下, 哥林多前/后书, 约翰一/二/三书) get a mark that
// tells them apart.
var glyph = map[string]string{
	"Gen": "创", "Exod": "出", "Lev": "利", "Num": "民", "Deut": "申",
	"Josh": "书", "Judg": "士", "Ruth": "得", "1Sam": "撒上", "2Sam": "撒下",
	"1Kgs": "列上", "2Kgs": "列下", "1Chr": "历上", "2Chr": "历下",
	"Ezra": "拉", "Neh": "尼", "Esth": "斯", "Job": "伯", "Ps": "诗",
	"Prov": "箴", "Eccl": "传", "Song": "歌", "Isa": "赛", "Jer": "耶",
	"Lam": "哀", "Ezek": "结", "Dan": "但", "Hos": "何", "Joel": "珥",
	"Amos": "摩", "Obad": "俄", "Jonah": "拿", "Mic": "弥", "Nah": "鸿",
	"Hab": "哈", "Zeph": "番", "Hag": "该", "Zech": "亚", "Mal": "玛",
	"Matt": "太", "Mark": "可", "Luke": "路", "John": "约", "Acts": "徒",
	"Rom": "罗", "1Cor": "林前", "2Cor": "林后", "Gal": "加", "Eph": "弗",
	"Phil": "腓", "Col": "西", "1Thess": "帖前", "2Thess": "帖后",
	"1Tim": "提前", "2Tim": "提后", "Titus": "多", "Phlm": "门",
	"Heb": "来", "Jas": "雅", "1Pet": "彼前", "2Pet": "彼后",
	"1John": "约一", "2John": "约二", "3John": "约三", "Jude": "犹", "Rev": "启",
}

// ZH is the book's Chinese name.
func ZH(osis string) string {
	if v, ok := chinese[osis]; ok {
		return v
	}
	return osis
}

// Mark is the short Chinese mark for a book tile.
func Mark(osis string) string {
	if v, ok := glyph[osis]; ok {
		return v
	}
	return osis
}
