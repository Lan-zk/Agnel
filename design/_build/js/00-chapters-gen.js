const CHAPTERS_GEN = {
  'gen-1': {
    book: 'gen', chapter: 1, label: '创世记 第 1 章',
    sections: [ { n: 1, from: 1, to: 31, titleEn: "The Creation of the World" } ],
    notes: {
      en: [ '1:26 "them": in the original Hebrew the verb is singular, referring to man as a whole (see footnote 3)', '1:27 "male and female": the original Hebrew reads "male" and "female", a pair of nouns (see footnote 4)' ],
      zh: [ '1:26 “他们”：原文的动词是单数，指整个人类。', '1:27 “造男造女”：原文是“男人”和“女人”，一对名词。' ]
    },
    verses: [
      { v: 1, en: 'In the beginning God created the heaven and the earth.', zh: '起初，神创造天地。' },
      { v: 2, en: 'And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters.', zh: '地是空虚混沌，渊面黑暗；神的灵运行在水面上。' },
      { v: 3, en: 'And God said, Let there be light: and there was light.', zh: '神说：要有光，就有了光。' },
      { v: 4, en: 'And God saw the light, that it was good: and God divided the light from the darkness.', zh: '神看光是好的，就把光暗分开了。' },
      { v: 5, en: 'And God called the light Day, and the darkness he called Night. And the evening and the morning were the first day.', zh: '神称光为昼，称暗为夜。有晚上，有早晨，这是头一日。' },
      { v: 6, en: 'And God said, Let there be a firmament in the midst of the waters, and let it divide the waters from the waters.', zh: '神说：诸水之间要有空气，将水分为上下。' },
      { v: 7, en: 'And God made the firmament, and divided the waters which were under the firmament from the waters which were above the firmament: and it was so.', zh: '神就造出空气，将空气以下的水、空气以上的水分开了。事就这样成了。' },
      { v: 8, en: 'And God called the firmament Heaven. And the evening and the morning were the second day.', zh: '神称空气为天。有晚上，有早晨，是第二日。' },
      { v: 9, en: 'And God said, Let the waters under the heaven be gathered together unto one place, and let the dry land appear: and it was so.', zh: '神说：天下的水要聚在一处，使旱地露出来。事就这样成了。' },
      { v: 10, en: 'And God called the dry land Earth; and the gathering together of the waters called he Seas: and God saw that it was good.', zh: '神称旱地为地，称水的聚处为海。神看着是好的。' },
      { v: 11, en: 'And God said, Let the earth bring forth grass, the herb yielding seed, and the fruit tree yielding fruit after his kind, whose seed is in itself, upon the earth: and it was so.', zh: '神说：地要发生青草和结种子的菜蔬，并结果子的树木，各从其类，果子都包着核。事就这样成了。' },
      { v: 12, en: 'And the earth brought forth grass, and herb yielding seed after his kind, and the tree yielding fruit, whose seed was in itself, after his kind: and God saw that it was good.', zh: '于是地发生了青草和结种子的菜蔬，各从其类；并结果子的树木，各从其类；果子都包着核。神看着是好的。' },
      { v: 13, en: 'And the evening and the morning were the third day.', zh: '有晚上，有早晨，是第三日。' },
      { v: 14, en: 'And God said, Let there be lights in the firmament of the heaven to divide the day from the night; and let them be for signs, and for seasons, and for days, and years:', zh: '神说：天上要有光体，可以分昼夜，作记号，定节令、日子、年岁，' },
      { v: 15, en: 'And let them be for lights in the firmament of the heaven to give light upon the earth: and it was so.', zh: '并要发光在天空，普照在地上。事就这样成了。' },
      { v: 16, en: 'And God made two great lights; the greater light to rule the day, and the lesser light to rule the night: he made the stars also.', zh: '于是神造了两个大光，大的管昼，小的管夜，又造众星，' },
      { v: 17, en: 'And God set them in the firmament of the heaven to give light upon the earth,', zh: '就把这些光摆列在天空，普照在地上，' },
      { v: 18, en: 'And to rule over the day and over the night, and to divide the light from the darkness: and God saw that it was good.', zh: '管理昼夜，分别明暗。神看着是好的。' },
      { v: 19, en: 'And the evening and the morning were the fourth day.', zh: '有晚上，有早晨，是第四日。' },
      { v: 20, en: 'And God said, Let the waters bring forth abundantly the moving creature that hath life, and fowl that may fly above the earth in the open firmament of heaven.', zh: '神说：水要多多滋生有生命的物；要有雀鸟飞在地面以上，天空之中。' },
      { v: 21, en: 'And God created great whales, and every living creature that moveth, which the waters brought forth abundantly, after their kind, and every winged fowl after his kind: and God saw that it was good.', zh: '神就造出大鱼和水中所滋生各样有生命的动物，各从其类；又造出各样飞鸟，各从其类。神看着是好的。' },
      { v: 22, en: 'And God blessed them, saying, Be fruitful, and multiply, and fill the waters in the seas, and let fowl multiply in the earth.', zh: '神就赐福给这一切，说：滋生繁多，充满海中的水；雀鸟也要多生在地上。' },
      { v: 23, en: 'And the evening and the morning were the fifth day.', zh: '有晚上，有早晨，是第五日。' },
      { v: 24, en: 'And God said, Let the earth bring forth the living creature after his kind, cattle, and creeping thing, and beast of the earth after his kind: and it was so.', zh: '神说：地要生出活物来，各从其类；牲畜、昆虫、地上的野兽，各从其类。事就这样成了。' },
      { v: 25, en: 'And God made the beast of the earth after his kind, and cattle after their kind, and every thing that creepeth upon the earth after his kind: and God saw that it was good.', zh: '于是神造出野兽，各从其类；牲畜，各从其类；地上一切昆虫，各从其类。神看着是好的。' },
      { v: 26, en: 'And God said, Let us make man in our image, after our likeness: and let them have dominion over the fish of the sea, and over the fowl of the air, and over the cattle, and over all the earth, and over every creeping thing that creepeth upon the earth.', zh: '神说：我们要照着我们的形像、按着我们的样式造人，使他们管理海里的鱼、空中的鸟、地上的牲畜，和全地，并地上所爬的一切昆虫。' },
      { v: 27, en: 'So God created man in his own image, in the image of God created he him; male and female created he them.', zh: '神就照着自己的形像造人，乃是照着他的形像造男造女。' },
      { v: 28, en: 'And God blessed them, and God said unto them, Be fruitful, and multiply, and replenish the earth, and subdue it: and have dominion over the fish of the sea, and over the fowl of the air, and over every living thing that moveth upon the earth.', zh: '神就赐福给他们，又对他们说：要生养众多，遍满地面，治理这地，也要管理海里的鱼、空中的鸟，和地上各样行动的活物。' },
      { v: 29, en: 'And God said, Behold, I have given you every herb bearing seed, which is upon the face of all the earth, and every tree, in the which is the fruit of a tree yielding seed; to you it shall be for meat.', zh: '神说：看哪，我将遍地上一切结种子的菜蔬和一切树上所结有核的果子全赐给你们作食物。' },
      { v: 30, en: 'And to every beast of the earth, and to every fowl of the air, and to every thing that creepeth upon the earth, wherein there is life, I have given every green herb for meat: and it was so.', zh: '至于地上的走兽和空中的飞鸟，并各样爬在地上有生命的物，我将青草赐给他们作食物。事就这样成了。' },
      { v: 31, en: 'And God saw every thing that he had made, and, behold, it was very good. And the evening and the morning were the sixth day.', zh: '神看着一切所造的都甚好。有晚上，有早晨，是第六日。' }
    ]
  },

  'gen-2': {
    book: 'gen', chapter: 2, label: '创世记 第 2 章',
    sections: [ { n: 2, from: 1, to: 3, titleEn: "Adam and Eve" }, { n: 3, from: 4, to: 8, titleEn: "The Garden of Eden" }, { n: 4, from: 9, to: 17, titleEn: "The Temptation and the Fall" }, { n: 5, from: 18, to: 20, titleEn: "Adam's Companionship" }, { n: 6, from: 21, to: 23, titleEn: "The Promise of the Redeemer" }, { n: 7, from: 24, to: 25, titleEn: "The Institution of Marriage" } ],
    notes: {
      en: [ '2:7 "the breath of life": the same Hebrew word is used for the Spirit in 1:2 (see footnote 5)', '2:25 "naked": the original Hebrew means "uncovered", the opposite of being ashamed (see footnote 6)' ],
      zh: [ '2:7 “生气”：原文所用的字，与 1:2 “神的灵”是同一个字。', '2:25 “赤身露体”：原文的意思是“没有遮蔽”，与羞耻相对。' ]
    },
    verses: [
      { v: 1, en: 'Thus the heavens and the earth were finished, and all the host of them.', zh: '天地万物都造齐了。' },
      { v: 2, en: 'And on the seventh day God ended his work which he had made; and he rested on the seventh day from all his work which he had made.', zh: '到第七日，神造物的工已经完毕，就在第七日歇了他一切的工，安息了。' },
      { v: 3, en: 'And God blessed the seventh day, and sanctified it: because that in it he had rested from all his work which God created and made.', zh: '神赐福给第七日，定为圣日；因为在这日，神歇了他一切创造的工，就安息了。' },
      { v: 4, en: 'These are the generations of the heavens and of the earth when they were created, in the day that the LORD God made the earth and the heavens,', zh: '创造天地的来历，在耶和华神造天地的日子，乃是这样，' },
      { v: 5, en: 'And every plant of the field before it was in the earth, and every herb of the field before it grew: for the LORD God had not caused it to rain upon the earth, and there was not a man to till the ground.', zh: '野地还没有草木，田间的菜蔬还没有长起来；因为耶和华神还没有降雨在地上，也没有人耕地，' },
      { v: 6, en: 'But there went up a mist from the earth, and watered the whole face of the ground.', zh: '但有雾气从地上腾，滋润遍地。' },
      { v: 7, en: 'And the LORD God formed man of the dust of the ground, and breathed into his nostrils the breath of life; and man became a living soul.', zh: '耶和华神用地上的尘土造人，将生气吹在他鼻孔里，他就成了有灵的活人，名叫亚当。' },
      { v: 8, en: 'And the LORD God planted a garden eastward in Eden; and there he put the man whom he had formed.', zh: '耶和华神在东方的伊甸立了一个园子，把所造的人安置在那里。' },
      { v: 9, en: 'And out of the ground made the LORD God to grow every tree that is pleasant to the sight, and good for food; the tree of life also in the midst of the garden, and the tree of knowledge of good and evil.', zh: '耶和华神使各样的树从地里长出来，可以悦人的眼目，其上的果子好作食物。园子当中又有生命树和分别善恶的树。' },
      { v: 10, en: 'And a river went out of Eden to water the garden; and from thence it was parted, and became into four heads.', zh: '有河从伊甸流出来，滋润那园子，从那里分为四道：' },
      { v: 11, en: 'The name of the first is Pison: that is it which compasseth the whole land of Havilah, where there is gold;', zh: '第一道名叫比逊，就是环绕哈腓拉全地的。在那里有金子，' },
      { v: 12, en: 'And the gold of that land is good: there is bdellium and the onyx stone.', zh: '并且那地的金子是好的；在那里又有珍珠和红玛瑙。' },
      { v: 13, en: 'And the name of the second river is Gihon: the same is it that compasseth the whole land of Ethiopia.', zh: '第二道河名叫基训，就是环绕古实全地的。' },
      { v: 14, en: 'And the name of the third river is Hiddekel: that is it which goeth toward the east of Assyria. And the fourth river is Euphrates.', zh: '第三道河名叫底格里斯，流在亚述的东边。第四道河就是伯拉河。' },
      { v: 15, en: 'And the LORD God took the man, and put him into the garden of Eden to dress it and to keep it.', zh: '耶和华神将那人安置在伊甸园，使他修理，看守。' },
      { v: 16, en: 'And the LORD God commanded the man, saying, Of every tree of the garden thou mayest freely eat:', zh: '耶和华神吩咐他说：园中各样树上的果子，你可以随意吃，' },
      { v: 17, en: 'But of the tree of the knowledge of good and evil, thou shalt not eat of it: for in the day that thou eatest thereof thou shalt surely die.', zh: '只是分别善恶树上的果子，你不可吃，因为你吃的日子必定死！' },
      { v: 18, en: 'And the LORD God said, It is not good that the man should be alone; I will make him an help meet for him.', zh: '耶和华神说：那人独居不好，我要为他造一个配偶帮助他。' },
      { v: 19, en: 'And out of the ground the LORD God formed every beast of the field, and every fowl of the air; and brought them unto Adam to see what he would call them: and whatsoever Adam called every living creature, that was the name thereof.', zh: '耶和华神用土所造成的野地各样走兽和空中各样飞鸟都带到那人面前，看他叫甚么。那人怎样叫各样的活物，那就是他的名字。' },
      { v: 20, en: 'And Adam gave names to all cattle, and to the fowl of the air, and to every beast of the field; but for Adam there was not found an help meet for him.', zh: '那人便给一切牲畜和空中飞鸟、野地走兽都起了名；只是那人没有遇见配偶帮助他。' },
      { v: 21, en: 'And the LORD God caused a deep sleep to fall upon Adam, and he slept: and he took one of his ribs, and closed up the flesh instead thereof;', zh: '耶和华神使他沉睡，他就睡了；于是取下他的一条肋骨，又把肉合起来。' },
      { v: 22, en: 'And the rib, which the LORD God had taken from man, made he a woman, and brought her unto the man.', zh: '耶和华神就用那人身上所取的肋骨造成一个女人，领他到那人跟前。' },
      { v: 23, en: 'And Adam said, This is now bone of my bones, and flesh of my flesh: she shall be called Woman, because she was taken out of Man.', zh: '那人说：这是我骨中的骨，肉中的肉，可以称他为女人，因为他是从男人身上取出来的。' },
      { v: 24, en: 'Therefore shall a man leave his father and his mother, and shall cleave unto his wife: and they shall be one flesh.', zh: '因此，人要离开父母，与妻子连合，二人成为一体。' },
      { v: 25, en: 'And they were both naked, the man and his wife, and were not ashamed.', zh: '当时夫妻二人赤身露体，并不羞耻。' }
    ]
  }
};
