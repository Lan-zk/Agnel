/* ===========================================================================
   数据层 —— 全部为演示用合成数据
   真实产品应替换为：BSB 英文圣经、ECDICT、CMU Pronouncing Dictionary、
   Open English WordNet（授权需逐项核验）。
   =========================================================================== */

const DEMO_CHAPTERS = {

  'gen-1': {
    book: 'gen', chapter: 1, label: '创世记 第 1 章',
    verses: [
      { v: 1, en: 'In the beginning God created the heavens and the earth.', zh: '起初，神创造天地。' },
      { v: 2, en: 'And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters.', zh: '地是混沌空虚的，深渊上黑暗覆盖；神的灵运行在水面上。' },
      { v: 3, en: 'And God said, Let there be light: and there was light.', zh: '神说，要有光，于是就有了光。' },
      { v: 4, en: 'And God saw the light, that it was good: and God divided the light from the darkness.', zh: '神看光是好的，就把光与暗分开了。' },
      { v: 5, en: 'And the evening and the morning were the first day.', zh: '晚上，早晨，是第一天。' }
    ]
  },
  'gen-2': {
    book: 'gen', chapter: 2, label: '创世记 第 2 章',
    verses: [
      { v: 1, en: 'Thus the heavens and the earth were finished, and all the host of them.', zh: '天地万象都造齐了。' },
      { v: 2, en: 'And on the seventh day God ended his work which he had made; and he rested on the seventh day from all his work which he had made.', zh: '到第七日，神造诸物就完毕了，就安息了。' },
      { v: 7, en: 'And the LORD God formed man of the dust of the ground, and breathed into his nostrils the breath of life; and man became a living soul.', zh: '耶和华神用地上的尘土造人，将生气吹在他鼻孔里，他就成了有灵的活人。' }
    ]
  },
  'ps-1': {
    book: 'ps', chapter: 1, label: '诗篇 第 1 篇',
    verses: [
      { v: 1, en: 'Blessed is the man that walketh not in the counsel of the ungodly.', zh: '不随从恶人的计谋、不行恶人的道路的人有福了。' },
      { v: 3, en: 'And he shall be like a tree planted by the rivers of water, that bringeth forth his fruit in his season; his leaf also shall not wither.', zh: '他好像栽种在溪水旁的树，按时候结果实；他的叶子也不枯干。' }
    ]
  },
  'ps-2': {
    book: 'ps', chapter: 2, label: '诗篇 第 2 篇',
    verses: [
      { v: 1, en: 'Why do the heathen rage, and the people imagine a vain thing?', zh: '外邦为什么喧嚷，列国为什么谋算虚空的事？' },
      { v: 4, en: 'They that dwell in the earth shall be astonished at his presence: the kings of the earth, and all the princes, shall gather themselves together.', zh: '住在地上的列国都要因他的同在而惊骇；地上的君王和首领要一同聚集。' }
    ]
  },
  'prov-1': {
    book: 'prov', chapter: 1, label: '箴言 第 1 章',
    verses: [
      { v: 1, en: 'The Proverbs of Solomon, the son of David, king of Israel.', zh: '以色列王大卫的儿子所罗门的箴言。' },
      { v: 7, en: 'The fear of the LORD is the beginning of knowledge: but fools despise wisdom and instruction.', zh: '敬畏耶和华是知识的开端；愚妄人藐视智慧和教训。' }
    ]
  },
  'matt-1': {
    book: 'matt', chapter: 1, label: '马太福音 第 1 章',
    verses: [
      { v: 12, en: 'And his father said, Wherefore then didst thou hide these three days from me?', zh: '他父亲说，那你为什么把这三天藏在我这里呢？' },
      { v: 20, en: 'While he was yet speaking to them, the angel of the Lord appeared to them in a dream, and said unto them, Fear not: for I bring you good tidings of great joy, which shall be to all people.', zh: '他正说着话，主的使者就在梦里显现给他们，说，不要怕，我报给你们大喜的信息，是给万民的。' }
    ]
  },
  'john-1': {
    book: 'john', chapter: 1, label: '约翰福音 第 1 章',
    verses: [
      { v: 1, en: 'In the beginning was the Word, and the Word was with God, and the Word was God.', zh: '太初有道，道与神同在，道就是神。' },
      { v: 3, en: 'All things were made by him; and without him was not any thing made that was made.', zh: '万物都是借着他造的，凡被造的，没有一样不是借着他造的。' },
      { v: 5, en: 'And the light shineth in the darkness; and the darkness hath not overcome it.', zh: '光照在黑暗里，黑暗不能把光胜了。' }
    ]
  },
  'john-2': {
    book: 'john', chapter: 2, label: '约翰福音 第 2 章',
    verses: [
      { v: 1, en: 'And the third day of the feast Jesus made a water of wine.', zh: '第三日娶筵的时候，耶稣用水变酒。' },
      { v: 6, en: 'Now there were set there six waterpots of stone, for the manner of the Jews, to cleanse in.', zh: '犹太人有用清水洗礼的规矩，所以这里摆着六口石缸。' }
    ]
  },
  'rom-1': {
    book: 'rom', chapter: 1, label: '罗马书 第 1 章',
    verses: [
      { v: 16, en: 'For I am not ashamed of the gospel of Christ: for it is the power of God unto salvation to every one that believeth.', zh: '我并不以福音为耻，因福音是神的大能，救一切信的人。' },
      { v: 17, en: 'For therein is the righteousness of God revealed from faith to faith: as it is written, The just shall live by faith.', zh: '因为福音的义是显明在人的信心里；就如经上记着：义人必因信得生。' }
    ]
  },
  'phil-1': {
    book: 'phil', chapter: 1, label: '腓立比书 第 1 章',
    verses: [
      { v: 6, en: 'He which hath begun a good work in you will also continue it unto the end of the day of Jesus Christ.', zh: '那在我心里动了善工的，必在新年完成，直到基督耶稣的日子。' },
      { v: 9, en: 'And this I pray, that your love may abound yet more and more in knowledge and in all judgement.', zh: '我所作的，就是叫你们的爱在知识和见识上多而又多。' }
    ]
  }
};

const CHAPTER_SETS = [
  typeof CHAPTERS_GEN !== 'undefined' ? CHAPTERS_GEN : null,
  typeof CHAPTERS_PSA !== 'undefined' ? CHAPTERS_PSA : null,
  typeof CHAPTERS_MTJN !== 'undefined' ? CHAPTERS_MTJN : null,
  typeof CHAPTERS_RMPP !== 'undefined' ? CHAPTERS_RMPP : null
].filter(Boolean);

const DATA = {
  books: [
    { id: 'gen',   zh: '创世记',   en: 'Genesis',     abbr: 'Gen',   part: '旧约', chapters: 2 },
    { id: 'ps',    zh: '诗篇',     en: 'Psalms',      abbr: 'Ps',    part: '旧约', chapters: 2 },
    { id: 'prov',  zh: '箴言',     en: 'Proverbs',    abbr: 'Prov',  part: '旧约', chapters: 1 },
    { id: 'matt',  zh: '马太福音', en: 'Matthew',     abbr: 'Matt',  part: '新约', chapters: 1 },
    { id: 'john',  zh: '约翰福音', en: 'John',        abbr: 'John',  part: '新约', chapters: 2 },
    { id: 'rom',   zh: '罗马书',   en: 'Romans',      abbr: 'Rom',   part: '新约', chapters: 1 },
    { id: 'phil',  zh: '腓立比书', en: 'Philippians', abbr: 'Phil',  part: '新约', chapters: 1 }
  ],

  // 每个经文单元：英文正文 + 中文对照
  chapters: Object.assign({}, DEMO_CHAPTERS, ...CHAPTER_SETS),

  // 普通英文文章（非经文），用于真实文本积累词汇
  articles: [
    {
      id: 'read-a-little',
      title: 'Reading a Little, Every Day',
      kicker: '学习笔记',
      paras: [
        {
          en: 'Most people fail at learning a language not because the method is wrong, but because they stop too early. A serious book asks for an hour; a serious vocabulary asks for fifteen minutes a day, repeated until the words stop feeling foreign.',
          zh: '多数人学不成语言，并非方法不对，而是坚持得太早退出。一本严肃的书需要一个小时；一份严肃的词汇量则需要每天十五分钟，反复练习，直到那些词不再陌生。'
        },
        {
          en: 'The useful unit of study is not the page but the sentence. When you meet a word in a sentence, you already have the context that gives it meaning, and the act of looking it up happens in the middle of a thought rather than beside it.',
          zh: '有用的学习单位不是页面，而是句子。当你在句子里遇到一个词，你已经拥有了赋予它意义的上下文，查词这个动作发生在思考之中，而不是思考之外。'
        },
        {
          en: 'Collecting a word without its sentence is like keeping a key with no door. The bank of words grows, but the memory of how each word behaved in a real line of English does not. So the sentence comes with the word, and the verse number comes with the sentence.',
          zh: '只收藏单词而不留句子，就像只留一把钥匙却没有门。词库在增长，但每个单词在真实英文行列中如何表现的记忆却没有增长。所以单词要带着句子，句子要带着经文编号。'
        },
        {
          en: 'Progress rarely arrives as a clean curve. It shows up as a morning when a word you met three weeks ago arrives without effort, before you have decided to search for it.',
          zh: '进步很少以干净的曲线到来。它表现为某个早晨：一个三周前遇到的词毫不费力地出现，在你还没决定去搜它之前。'
        },
        {
          en: 'What you do in that moment decides how many such mornings you get. Open the book, read the next sentence, and keep going. The fifteen minutes are enough; the rest is only whether you start.',
          zh: '你此刻怎么做，决定了你会有多少个这样的早晨。打开书，读下一句，继续。十五分钟就够了；剩下的只是你会不会开始。'
        }
      ]
    }
  ],

  // 英汉词典：音标、词性、释义、词形变化、派生、搭配、词源
  dict: {
    create: {
      ipa: '/kriˈeɪt/', ipaUS: '/kriˈeɪt/', pos: 'v.',
      senses: ['创造；创作', '造成（局面、结果）', '生（小孩）'],
      forms: ['create', 'creates', 'created', 'creating'],
      derivations: ['creation', 'creator', 'creative', 'creativity'],
      etymology: '拉丁语 creāre（生长、造成）',
      collocations: ['create a world', 'create a habit', 'create an account'],
      root: 'cre'
    },
    created: {
      ipa: '/kriˈeɪtɪd/', pos: 'v. create 的过去式 / 过去分词',
      senses: ['创造了', '造成了'],
      forms: ['create', 'creates', 'created', 'creating'],
      derivations: ['creation', 'creator'],
      etymology: '同源：create',
      collocations: ['God created', 'was created for'],
      root: 'cre'
    },
    beginning: {
      ipa: '/bɪˈɡɪnɪŋ/', pos: 'n.',
      senses: ['开始；开端', '起源'],
      forms: ['begin', 'begins', 'began', 'begun', 'beginning'],
      derivations: ['beginner'],
      etymology: '古英语 beginnan',
      collocations: ['from the beginning', 'a new beginning'],
      root: 'begin'
    },
    light: {
      ipa: '/laɪt/', pos: 'n. / adj.',
      senses: ['n. 光；灯光；光明', 'adj. 轻的；明亮的'],
      forms: ['light', 'lights', 'lighted', 'lighting'],
      derivations: ['lighten', 'lighting', 'lightning', 'delight'],
      etymology: '古英语 lēoht',
      collocations: ['let there be light', 'the light of the world'],
      root: 'light'
    },
    darkness: {
      ipa: '/ˈdɑːk.nəs/', pos: 'n.',
      senses: ['黑暗；暗处', '无知；愚昧'],
      forms: ['dark', 'darkness'],
      derivations: ['darken'],
      etymology: '古英语 deorc',
      collocations: ['in the darkness', 'the darkness cannot overcome it'],
      root: 'dark'
    },
    without: {
      ipa: '/wɪˈðaʊt/', pos: 'prep. / adv.',
      senses: ['prep. 没有；缺少', 'adv. 不；向外'],
      forms: ['without'],
      derivations: [],
      etymology: '古英语 wītan',
      collocations: ['without form', 'without fear'],
      root: 'with'
    },
    trust: {
      ipa: '/trʌst/', pos: 'v. / n.',
      senses: ['信任；信赖', '信托；信托财产'],
      forms: ['trust', 'trusts', 'trusted', 'trusting'],
      derivations: ['trusted', 'trustworthy'],
      etymology: '古英语 trēst',
      collocations: ['trust in the LORD', 'put your trust in'],
      root: 'trust'
    },
    understanding: {
      ipa: '/ˌʌndəˈstændɪŋ/', pos: 'n.',
      senses: ['理解；领会', '理解力', '谅解'],
      forms: ['understand', 'understands', 'understood', 'understanding'],
      derivations: ['understandable', 'misunderstanding'],
      etymology: 'understand + -ing（名词后缀）',
      collocations: ['all thine heart', 'a shared understanding'],
      root: 'stand'
    },
    shadow: {
      ipa: '/ˈʃæd.oʊ/', pos: 'n.',
      senses: ['影子；阴影', '庇护；荫蔽'],
      forms: ['shadow', 'shadows', 'shadowed', 'shadowing'],
      derivations: ['shadowy'],
      etymology: '古英语 sceadu',
      collocations: ['the shadow of death', 'in shadow'],
      root: 'shade'
    },
    death: {
      ipa: '/deθ/', pos: 'n.',
      senses: ['死亡', '灭亡；终结'],
      forms: ['die', 'dies', 'died', 'dying', 'death'],
      derivations: ['deadly', 'deathless'],
      etymology: '古英文 dēath',
      collocations: ['the valley of the shadow of death'],
      root: 'die'
    },
    comfort: {
      ipa: '/ˈkʌm.fət/', pos: 'v. / n.',
      senses: ['安慰；使舒适', '舒适；安慰物'],
      forms: ['comfort', 'comforts', 'comforted', 'comforting'],
      derivations: ['comfortable', 'comfortably'],
      etymology: '拉丁语 confortare（加强）',
      collocations: ['they comfort me', 'a comfort to'],
      root: 'fort'
    },
    word: {
      ipa: '/wɜːd/', pos: 'n.',
      senses: ['词；单词', '话；言语', '消息'],
      forms: ['word', 'words', 'worded', 'wording'],
      derivations: ['wordy', 'wordless', 'wording'],
      etymology: '古英语 word',
      collocations: ['the Word of God', 'in a word'],
      root: 'word'
    },
    transform: {
      ipa: '/trænsˈfɔːm/', pos: 'v.',
      senses: ['改变；转变', '改造；变革'],
      forms: ['transform', 'transforms', 'transformed', 'transforming'],
      derivations: ['transformation', 'transformer'],
      etymology: '拉丁语 transformāre（trans- 越过 + forma 形状）',
      collocations: ['transform your mind', 'be transformed by'],
      root: 'form'
    },
    strengthen: {
      ipa: '/stren.θən/', pos: 'v.',
      senses: ['加强；增强', '使强壮；巩固'],
      forms: ['strengthen', 'strengthens', 'strengthened', 'strengthening'],
      derivations: ['strength', 'strong', 'strongly'],
      etymology: 'strong + -en（动词后缀）',
      collocations: ['strengtheneth me', 'strengthen your faith'],
      root: 'strong'
    },
    walk: {
      ipa: '/wɔːk/', pos: 'v. / n.',
      senses: ['走；步行', '散步；走路'],
      forms: ['walk', 'walks', 'walked', 'walking'],
      derivations: ['walker'],
      etymology: '古英语 wealcan',
      collocations: ['walk through', 'walk with God'],
      root: 'walk'
    },
    harvest: {
      ipa: '/ˈhɑː.r.vəst/', pos: 'n. / v.',
      senses: ['收获；收成', '收割'],
      forms: ['harvest', 'harvests', 'harvested', 'harvesting'],
      derivations: ['harvester'],
      etymology: '古英语 hærfest',
      collocations: ['bring forth his fruit', 'the harvest is ripe'],
      root: 'harvest'
    },
    plant: {
      ipa: '/plænt/', pos: 'v. / n.',
      senses: ['种植；栽', '植物；草木'],
      forms: ['plant', 'plants', 'planted', 'planting'],
      derivations: ['planting', 'plantation'],
      etymology: '拉丁语 planta',
      collocations: ['planted by the rivers', 'plant a tree'],
      root: 'plant'
    },
    river: {
      ipa: '/ˈrɪv.ər/', pos: 'n.',
      senses: ['河；江', '河流'],
      forms: ['river', 'rivers'],
      derivations: ['rival'],
      etymology: '拉丁语 rivus',
      collocations: ['by the rivers of water', 'the river flows'],
      root: 'riv'
    },
    verse: {
      ipa: '/vɜːs/', pos: 'n.',
      senses: ['（圣经的）节', '诗节；韵文', '诗句'],
      forms: ['verse', 'verses'],
      derivations: ['versify', 'version'],
      etymology: '拉丁语 versus（转动的）',
      collocations: ['read a verse', 'in this verse'],
      root: 'vert'
    },
    gather: {
      ipa: '/ˈɡæð.ər/', pos: 'v.',
      senses: ['聚集；集合', '收集；采集'],
      forms: ['gather', 'gathers', 'gathered', 'gathering'],
      derivations: ['gathering'],
      etymology: '古英语 gaderian',
      collocations: ['gather themselves', 'gather the words'],
      root: 'gad'
    }
  },

  // 词根与派生：严格区分 词形变化 / 派生关系 / 真实词源
  roots: [
    {
      root: 'cre',
      gloss: '生长、造成',
      origin: '拉丁语 creāre',
      members: {
        forms: ['create', 'creates', 'created', 'creating'],
        derivations: [
          { word: 'creation', gloss: '创造；作品' },
          { word: 'creator', gloss: '创造者' },
          { word: 'creative', gloss: '有创造力的' },
          { word: 'creativity', gloss: '创造力' }
        ]
      }
    },
    {
      root: 'strong',
      gloss: '强',
      origin: '古英语 strong',
      members: {
        forms: ['strong', 'stronger', 'strongest', 'strengthened'],
        derivations: [
          { word: 'strength', gloss: '力量；强度' },
          { word: 'strengthen', gloss: '加强' },
          { word: 'strongly', gloss: '强烈地' }
        ]
      }
    },
    {
      root: 'form',
      gloss: '形状',
      origin: '拉丁语 forma',
      members: {
        forms: ['transform', 'transforms', 'transformed'],
        derivations: [
          { word: 'transformation', gloss: '变革；转变' },
          { word: 'transformer', gloss: '变压器' },
          { word: 'reform', gloss: '改革' }
        ]
      }
    },
    {
      root: 'stand',
      gloss: '站立',
      origin: '古英语 standan',
      members: {
        forms: ['understand', 'understood', 'understanding'],
        derivations: [
          { word: 'understandable', gloss: '可理解的' },
          { word: 'misunderstanding', gloss: '误解' }
        ]
      }
    },
    {
      root: 'light',
      gloss: '光；轻',
      origin: '古英语 lēoht',
      members: {
        forms: ['light', 'lights', 'lighted'],
        derivations: [
          { word: 'lighten', gloss: '减轻；照亮' },
          { word: 'lightning', gloss: '闪电' },
          { word: 'delight', gloss: '喜悦' }
        ]
      }
    }
  ],

  // 语法句法：例句全部取自上面的阅读语料
  grammar: [
    {
      point: 'there be 表示"存在"',
      note: 'be 动词跟 there 连用，表示某处存在某物；单数用 is，复数用 are，过去用 was/were。',
      en: 'And there was light.',
      zh: '于是就有了光。',
      from: '创世记 1:3'
    },
    {
      point: '被动语态：动作的承受者在前',
      note: 'be + 过去分词。经文里常见 "be made / be divided / be hidden"，动作的施动者常被省略。',
      en: 'All things were made by him.',
      zh: '万物都是借着他造的。',
      from: '约翰福音 1:3'
    },
    {
      point: '关系从句：that / which 引导',
      note: 'that 指人指物皆可，which 只指物。修饰前面的名词，让长句分成两半更好读。',
      en: 'The Proverbs of Solomon, the son of David, king of Israel.',
      zh: '以色列王大卫的儿子所罗门的箴言。',
      from: '箴言 1:1'
    },
    {
      point: '现在完成时：过去的动作影响到现在',
      note: 'have / has + 过去分词。常与 already、yet、since、for 搭配。',
      en: 'He which hath begun a good work in you will also continue it.',
      zh: '那在我心里动了善工的，必完成。',
      from: '腓立比书 1:6'
    },
    {
      point: '祈使句：说话人或真理的直接指令',
      note: '省掉主语，直接用动词原形开头。经文里 "Let there be" 是最典型的一例。',
      en: 'And God said, Let there be light.',
      zh: '神说，要有光。',
      from: '创世记 1:3'
    },
    {
      point: '分词短语作状语：with / without + 名词/代词 + 分词',
      note: 'with 后面接"伴随的东西"，without 后面接"缺少的东西"。这是英文里非常紧凑的一种表达。',
      en: 'And the earth was without form, and void.',
      zh: '地是混沌空虚的。',
      from: '创世记 1:2'
    },
    {
      point: '动名词与不定式：介词后一律用动名词',
      note: 'by + doing 表示"通过做某事"；to do 表示目的。看清介词就能选对形式。',
      en: 'For it is the power of God unto salvation to every one that believeth.',
      zh: '因福音是神的大能，救一切信的人。',
      from: '罗马书 1:16'
    },
    {
      point: '比较级与最高级',
      note: '单音节加 -er / -est，多音节用 more / most；than 后接比较对象。',
      en: 'That is good: and God divided the light from the darkness.',
      zh: '神看光是好的。',
      from: '创世记 1:4'
    }
  ],

  // 音标：元音与辅音
  phonics: {
    vowels: [
      { ipa: 'iː', sample: 'sheep', zh: '长音 i', hint: '嘴角向两边拉，舌位偏高' },
      { ipa: 'ɪ', sample: 'ship', zh: '短音 i', hint: '短促放松，不要拖长' },
      { ipa: 'e', sample: 'bed', zh: '短音 e', hint: '嘴角略开，舌位中前' },
      { ipa: 'æ', sample: 'cat', zh: '扁 a', hint: '下巴放平，嘴张大一点' },
      { ipa: 'ɑː', sample: 'car', zh: '长音 a', hint: '舌位靠后，声音从喉部送出' },
      { ipa: 'ɒ', sample: 'dog', zh: '短音 o', hint: '双唇收圆，舌位后缩' },
      { ipa: 'ɔː', sample: 'law', zh: '长音 o', hint: '双唇更圆，拖长' },
      { ipa: 'ʊ', sample: 'book', zh: '短音 u', hint: '嘴唇放松，不要用力' },
      { ipa: 'uː', sample: 'moon', zh: '长音 u', hint: '双唇撮圆前突' },
      { ipa: 'ʌ', sample: 'cup', zh: '央元音', hint: '口腔中央，短而有力' },
      { ipa: 'ɜː', sample: 'bird', zh: '长音 er', hint: '舌位居中，双唇扁平' },
      { ipa: 'ə', sample: 'about', zh: '弱读 schwa', zh2: '', hint: '英语里最常见的弱读音，轻、快、含糊' }
    ],
    consonants: [
      { ipa: 'p', sample: 'pen', zh: '清辅音 p', hint: '送气，词首要加气' },
      { ipa: 'b', sample: 'book', zh: '浊辅音 b', hint: '声带振动' },
      { ipa: 't', sample: 'ten', zh: '清辅音 t', hint: '舌尖抵上齿龈' },
      { ipa: 'd', sample: 'door', zh: '浊辅音 d', hint: '同部位，声带振动' },
      { ipa: 'k', sample: 'cat', zh: '清辅音 k', hint: '舌根抵软腭' },
      { ipa: 'g', sample: 'go', zh: '浊辅音 g', hint: '同部位，声带振动' },
      { ipa: 'f', sample: 'five', zh: '清辅音 f', hint: '上齿咬下唇' },
      { ipa: 'v', sample: 'very', zh: '浊辅音 v', hint: '上齿咬下唇 + 声带' },
      { ipa: 'θ', sample: 'think', zh: '齿间音', hint: '舌尖轻放在上下齿之间' },
      { ipa: 'ð', sample: 'this', zh: '浊齿间音', hint: '词尾的 the 常是这个音' },
      { ipa: 's', sample: 'sun', zh: '清辅音 s', hint: '像蛇的声音' },
      { ipa: 'z', sample: 'zoo', zh: '浊辅音 z', hint: '同 s，声带振动' },
      { ipa: 'ʃ', sample: 'she', zh: '摩擦音 sh', hint: '双唇略前噘' },
      { ipa: 'ʒ', sample: 'measure', zh: '浊摩擦音', hint: 'sh + 声带振动' },
      { ipa: 'tʃ', sample: 'church', zh: '塞擦音 ch', hint: '先塞后擦' },
      { ipa: 'dʒ', sample: 'judge', zh: '浊塞擦音', hint: 'ch + 声带振动' },
      { ipa: 'm', sample: 'moon', zh: '鼻音 m', hint: '气流从鼻腔出来' },
      { ipa: 'n', sample: 'no', zh: '鼻音 n', hint: '同 m，舌位靠后' },
      { ipa: 'ŋ', sample: 'sing', zh: '后鼻音 ng', hint: '舌根抵住软腭' },
      { ipa: 'h', sample: 'hand', zh: '声门擦音', hint: '只是轻轻送气' },
      { ipa: 'r', sample: 'red', zh: '卷舌 r', hint: '舌尖不碰上颚，轻轻卷起' },
      { ipa: 'l', sample: 'light', zh: '边音 l', hint: '词首舌尖抵齿龈，词尾双唇略圆' },
      { ipa: 'w', sample: 'we', zh: '半元音 w', hint: '双唇收圆后迅速滑开' },
      { ipa: 'j', sample: 'yes', zh: '半元音 y', hint: '舌面抬向硬腭后滑开' }
    ]
  },

  // 基础词汇与短语
  basics: {
    verbs: [
      { w: 'create', zh: '创造' }, { w: 'trust', zh: '信任' }, { w: 'read', zh: '读' },
      { w: 'remember', zh: '记住' }, { w: 'begin', zh: '开始' }, { w: 'seek', zh: '寻求' },
      { w: 'walk', zh: '行走' }, { w: 'speak', zh: '说话' }, { w: 'keep', zh: '保持' },
      { w: 'gather', zh: '聚集' }, { w: 'plant', zh: '种植' }, { w: 'shine', zh: '照耀' }
    ],
    nouns: [
      { w: 'light', zh: '光' }, { w: 'word', zh: '道；话语' }, { w: 'beginning', zh: '开端' },
      { w: 'shadow', zh: '影子' }, { w: 'strength', zh: '力量' }, { w: 'river', zh: '河' },
      { w: 'verse', zh: '经节' }, { w: 'harvest', zh: '收成' }, { w: 'spirit', zh: '灵' },
      { w: 'voice', zh: '声音' }, { w: 'habit', zh: '习惯' }, { w: 'memory', zh: '记忆' }
    ],
    phrases: [
      { w: 'in the beginning', zh: '起初' }, { w: 'without form', zh: '没有形状' },
      { w: 'the face of the deep', zh: '深渊的表面' }, { w: 'let there be light', zh: '要有光' },
      { w: 'in his season', zh: '按时候' }, { w: 'the fear of the LORD', zh: '敬畏耶和华' },
      { w: 'the beginning of knowledge', zh: '知识的开端' }, { w: 'without fear', zh: '不惧' },
      { w: 'by faith', zh: '因信' }, { w: 'unto salvation', zh: '以致得救' },
      { w: 'a good work', zh: '善工' }, { w: 'every one that believeth', zh: '一切信的人' }
    ]
  }
};

/* 演示词条：第一次打开时预置几条，方便看到词库与练习的样子。
   全部取自下面的正文，标记 demo，界面上会写明是演示数据。 */
const DEMO_BANK = [
  { word: 'created', stage: 'new', ctx: { source: '创世记 1:1', en: 'In the beginning God created the heavens and the earth.', zh: '起初，神创造天地。' } },
  { word: 'beginning', stage: 'new', ctx: { source: '创世记 1:1', en: 'In the beginning God created the heavens and the earth.', zh: '起初，神创造天地。' } },
  { word: 'without', stage: 'learning', ctx: { source: '创世记 1:2', en: 'And the earth was without form, and void; and darkness was upon the face of the deep.', zh: '地是混沌空虚的，深渊上黑暗覆盖。' } },
  { word: 'light', stage: 'learning', ctx: { source: '创世记 1:3', en: 'And God said, Let there be light: and there was light.', zh: '神说，要有光，于是就有了光。' } },
  { word: 'comfort', stage: 'new', ctx: { source: '诗篇 23:4', en: 'Thy rod and thy staff they comfort me.', zh: '你的杖和你的竿都安慰我。' } },
  { word: 'trust', stage: 'known', ctx: { source: '诗篇 23:1', en: 'The LORD is my shepherd; I shall not want.', zh: '耶和华是我的牧者，我必不至缺乏。' } }
];
