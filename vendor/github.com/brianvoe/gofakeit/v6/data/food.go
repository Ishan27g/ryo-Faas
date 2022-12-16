package data

// Food consists of food information
var Food = map[string][]string{
	"fruit":     {"Apple", "Apricot", "Avocado", "Banana", "Bilberry", "Blackberry", "Blackcurrant", "Blueberry", "Currant", "Cherry", "Cherimoya", "Clementine", "Date", "Damson", "Durian", "Eggplant", "Elderberry", "Feijoa", "Gooseberry", "Grape", "Grapefruit", "Guava", "Huckleberry", "Jackfruit", "Jambul", "Kiwi", "Kumquat", "Legume", "Lemon", "Lime", "Lychee", "Mango", "Mangostine", "Melon", "Cantaloupe", "Honeydew", "Watermelon", "Rock melon", "Nectarine", "Orange", "Peach", "Pear", "Pitaya", "Physalis", "Plum", "Pineapple", "Pomegranate", "Raisin", "Raspberry", "Rambutan", "Redcurrant", "Satsuma", "Strawberry", "Tangerine", "Tomato", "Watermelon"},
	"vegetable": {"Amaranth Leaves", "Arrowroot", "Artichoke", "Arugula", "Asparagus", "Bamboo Shoots", "Beans, Green", "Beets", "Belgian Endive", "Bitter Melon*", "Bok Choy", "Broadbeans", "Broccoli", "Broccoli Rabe", "Brussel Sprouts", "Cabbage", "Carrot", "Cassava", "Cauliflower", "Celeriac", "Celery", "Chicory", "Collards", "Corn", "Crookneck", "Cucumber", "Daikon", "Dandelion Greens", "Eggplant", "Fennel", "Fiddleheads", "Ginger Root", "Horseradish", "Jicama", "Kale", "Kohlrabi", "Leeks", "Lettuce", "Mushrooms", "Mustard Greens", "Okra", "Onion", "Parsnip", "Peas", "Pepper", "Potato", "Pumpkin", "Radicchio", "Radishes", "Rutabaga", "Salsify", "Shallots", "Snow Peas", "Sorrel", "Soybeans", "Spaghetti Squash", "Spinach", "Squash", "Sugar Snap Peas", "Sweet Potato", "Swiss Chard", "Tomato", "Turnip", "Watercress", "Yam Root", "Zucchini"},
	"breakfast": {"berry cream cheese coffee cake", "broiled cinnamon toast", "breakfast casserole seasoned with country gravy", "mamas fruit cobbler", "shirleys plain or blueberry muffins", "toasted sunny side up egg and cheese sandwiches", "3 meat breakfast pizza", "moms cheat doughnuts", "old fashioned banana muffins", "blackberry breakfast bars", "pikelets australian pancakes", "pumpkin ginger scones with cinnamon chips", "tomato and mushroom omelette", "asparagus omelette wraps", "poached eggs technique", "scrambled egg sandwiches with onions and red peppers", "cheesecake kugel", "chicken and egg on rice oyako donburi", "bacon egg casserole", "ginger lemon muffins", "lizs morning glory muffins", "scrambled eggs oeufs brouills", "nats cucumber cream cheese bagel", "easy breakfast casserole", "6 week bran muffins auntie annes muffins", "awesome orange chocolate muffins", "baked swiss cheese omelet", "melt in your mouth blueberry muffins", "baked pears", "flaeskeaeggekage danish bacon egg pancake omelet", "sleepy twisted sisters g n g breakfast ramekin", "lemon buttercream pancakes with blueberries", "chef flowers simple sunday brunch omelette", "blueberry bakery muffins", "cardamom sour cream waffles", "sausage gravy for biscuits and gravy", "creamy scrambled eggs in the microwave", "english muffins with bacon butter", "original praline bacon recipe", "christmas caramel rolls easy", "blueberry banana happy face pancakes", "whole grain pancake mix", "fresh mango bread", "canadian bacon cheese omelet", "pumpkin french toast with toasted walnuts", "green mountain granola", "italian eggs with bacon", "a faster egg muffin", "country scrambled eggs", "everyday french breakfast baguette and jam with chocolate milk", "mexi eggs in a hole", "fruited irish oatmeal", "ham omelet deluxe", "danish bubble", "best buttermilk pancakes", "egg flowers", "vanilla fruit dip", "eggs in a basket", "grandmas swedish thin pancakes", "cinnamon maple granola", "wake up stuffed french breakfast panini", "quinoa muffins", "grilled cheese on raisin bread", "castillian hot chocolate", "banana blueberry oatmeal bread", "caramel pull aparts", "purple cow", "chili jack oven omelet", "cheery cherry muffins", "israeli breakfast salad", "muffin toppings", "migas lite for 2", "easy danish kringle", "oatmeal cookie granola"},
	"lunch":     {"no bake hersheys bar pie", "worm sandwiches", "quesadillas for one or two", "pearls sesame noodles", "patty melt", "fresh tomato sandwiches saturday lunch on longmeadow farm", "onion burgers by john t edge the longmeadow farm", "fresh tomato and cucumber salad", "hoisin marinated wing pieces", "feta marinated", "spicy roasted butternut seeds pumpkin seeds", "honey chipotle pecans", "baked ham glazed with pineapple and chipotle peppers", "reuben sandwich our way", "toasted sunny side up egg and cheese sandwiches", "mrs allens date loaf", "3 meat breakfast pizza", "body and soul health muffins", "grilled blue cheese burgers", "kittencals beef burritos", "spinach and mandarin orange salad", "coconut pound cake", "scallop saute", "open faced crab sandwiches", "the traditional cyprus sandwich with halloumi onions and tomato", "toasted ham and cheese supreme", "scrambled egg sandwiches with onions and red peppers", "cucumber open faced sandwiches", "chicken and egg on rice oyako donburi", "blt sandwich", "grilled chicken pesto panini", "mushroom and chicken grilled quesadillas", "delicious cheesy bacon and green onion potato skins", "grilled chili lime chicken", "fried almonds", "the greatful bread sandwich", "egg salad club sandwiches or shrimp salad club", "nifs peanut butter banana muffins", "parmesan fish in the oven", "caramelized onion focaccia bread machine", "nats cucumber cream cheese bagel", "chicken with cashews", "lemon parsley popcorn", "not your ordinary chocolate chip cookies liqueur laced", "katos tasty salmon cream cheese surprise", "greek inspired salad", "tomato basil american cheese sandwich", "club sandwich", "bacon and egg salad sandwiches", "apple cheese bites", "two cheese panini with tomato olive pesto", "delicious and simple fruit dip", "tex mex 7 layer salad", "grilled peanut butter and jelly sandwich", "simply simple cucumber slices in vinegar dressing longmeadow", "ww greek inspired scrambled egg wraps", "baby greens with mustard vinaigrette", "patty melts", "ribs", "chocolate angel food cake", "spinach with lemon garlic", "green goddess dressing", "leftover rice muffins", "cajun garlic fingers", "fresh mango bread", "california crab salad", "hot salty nuts", "beef for tacos", "hidden valley wraps", "omas boterkoek dutch buttercake", "apple butterflies", "don t burn your fingers garlic bread", "beer wisconsin bratwurst", "salmon with bourbon and brown sugar glaze", "lemon coconut muffins", "the godfather of grilled cheese sandwiches", "green mountain granola", "tuna red onion and parsley salad", "tortellini skewers", "italian meatball hoagies", "crispy fried chicken spring rolls", "rotisserie style chicken in the crock pot", "creamed peas on toast", "bergy dim sum 5 steamed shrimp dumplings", "chocolate almond roca bar", "number 400 seafood casserole", "chocolate rainbow krispies treats", "spinach salad with blue cheese", "hash", "fake crab salad sandwiches", "guacamole stuffed deviled eggs", "weight watchers veggie barley soup 1 pt for 1 cup", "hummus with a twist", "bellissimo panini", "carls jr western bacon cheeseburger copycat by todd wilbur", "salami havarti and cole slaw sandwiches", "garlic herbed roasted red skin potatoes", "grilled cheese on raisin bread", "hearty grilled cheese", "italian deli wraps", "strammer max german warm sandwich", "quick elephant ears", "salata marouli romaine lettuce salad", "goat cheese black olive mashed potatoes", "tomato cucumber avocado sandwich", "purple cow", "chocolate coconut dream bars", "homemade popsicles", "ginger soy salmon", "sweet and sour pork balls", "spicy chicken soup with hints of lemongrass and coconut milk", "another buffalo wings recipe", "famous white wings", "amazing sweet italian sausage pasta soup", "sausage sandwich italian style", "copycat taco bell chicken enchilada bowl", "simple pan fried chicken breasts", "1 2 3 black bean salsa dip", "quick chile relleno casserole", "bacon spaghetti squash", "fantastic banana bran muffins", "garbanzo vegetarian burgers", "mediterranean tuna stuffed tomato", "sugared cinnamon almonds", "queen margherita pizza", "insanely easy chickpea salad", "habit forming shrimp dip", "turkey swiss panini", "pumpkin chocolate chip muffins", "grilled havarti and avocado sandwiches", "english muffin pizzas", "oatmeal cookie granola"},
	"dinner":    {"kittencals caesar tortellini salad", "no bake hersheys bar pie", "lindas special potato salad", "kittencals parmesan orzo", "pearls sesame noodles", "roasted potatoes and green beans", "kittencals really great old fashioned lemonade", "lindas chunky garlic mashed potatoes", "kittencals pan fried asparagus", "cafe mocha latte", "fresh tomato and cucumber salad", "peanut butter gooey cake", "foolproof standing prime rib roast paula deen", "mamas fruit cobbler", "hoisin marinated wing pieces", "feta marinated", "the realtors cream cheese corn", "savory pita chips", "jalapeno pepper jelly chicken", "kashmir lamb with spinach", "oven fried zucchini sticks", "best ever bruschetta", "maple cinnamon coffee", "kick a fried onion rings", "guava mojito", "confit d oignon french onion marmalade", "flounder stuffed with shrimp and crabmeat", "mrs allens date loaf", "swedish cucumber salad pressgurka", "authentic pork lo mein chinese", "golden five spice sticky chicken", "basil tomato salad", "white chocolate cheesecake", "celery and blue cheese salad", "kittencals crock pot french dip roast", "lindas asian salmon", "spinach and mandarin orange salad", "coconut pound cake", "scallop saute", "spicy catfish tenders with cajun tartar sauce", "just like deweys candied walnut and grape salad", "strawberry pavlova", "grilled pork chops with lime cilantro garlic", "smoky barbecue beef brisket crock pot", "quick and easy chicken in cream sauce", "fried chorizo with garlic", "cucumber open faced sandwiches", "rachael rays mimosa", "tortellini bow tie pasta salad", "tonkatsu japanese pork cutlet", "mushroom and chicken grilled quesadillas", "delicious cheesy bacon and green onion potato skins", "roasted beet salad with horseradish cream dressing", "islands bananas foster", "apricot glazed roasted asparagus low fat", "frozen kahlua creme", "fried almonds", "just peachy grillin ribs rsc", "death by chocolate cake", "parmesan fish in the oven", "calico peas", "creamy cucumber dill dip", "emerils stewed black eyed peas", "german style eiskaffee iced coffee drink", "strawberry angel trifle", "spinach salad with feta cheese", "french napoleons", "ultimate crab and spinach manicotti with parmesan cheese sauce", "sweet and sour stir fry shrimp with broccoli and red bell pepper", "crispy noodle salad with sweet and sour dressing", "crunchy rosemary potatoes", "roasted cherry or grape tomatoes", "blackened skillet shrimp", "parslied new potatoes", "tropical baked chicken", "sweet and sour kielbasa kabobs", "fantastic mushrooms with garlic butter and parmesan", "asparagus with lemon butter crumbs", "creamy garlic prawns", "kittencals banana almond muffins with almond streusel", "ww shrimp scampi", "kittencals tender microwave corn with husks on", "nude beach", "kittencals greek garden salad with greek style dressing", "roasted broccoli with cherry tomatoes", "kittencals chicken cacciatore", "buttermilk mashed potatoes with country mustard", "tilapia in thai sauce", "cream cheese potato soup", "brown sugar roasted salmon with maple mustard dill sauce", "baby greens with mustard vinaigrette", "ribs", "new england roasted cornish game hens", "chocolate angel food cake", "creamy strawberries", "spinach with lemon garlic", "green goddess dressing", "jamaican pork tenderloin", "awesome twice baked potatoes", "sausage mushroom appetizers", "roasted garlic soup with parmesan", "crushed red potatoes with garlic", "15 minute no fry chicken enchiladas honest", "uncle bills caesar canadian style", "raspberry cranberry salad with sour cream cream cheese topping", "hot salty nuts", "acorn squash for 2", "pumpkin knot yeast rolls", "caramelized onion dip spread", "roasted asparagus with sage and lemon butter", "spanish garlic shrimp taverna", "baby greens with pears gorgonzola and pecans", "grilled or baked salmon with lavender", "ruth walls german apple cake", "healthy italian breadsticks or pizza crust", "strawberry and cream cheese parfait", "marinated grilled tuna steak", "kittencals extra crispy fried chicken breast", "de constructed chicken cordon bleu", "moroccan cinnamon coffee with orange flower water", "lemon and parsley potatoes", "bergy dim sum 5 steamed shrimp dumplings", "chocolate almond roca bar", "garlic mashed potatoes and cashew gravy", "number 400 seafood casserole", "sherry buttered shrimp", "spinach salad with blue cheese", "cookie monster fruit salad", "asian broccoli salad", "pink poodle", "butterflied leg of lamb with lots of garlic and rosemary", "gorgonzola and toasted walnut salad", "maple coffee", "chocolate chip bundt cake with chocolate glaze", "crock pot caramelized onion pot roast", "mashed potatoes with bacon and cheddar", "provencal olives", "creole potato salad", "wild addicting dip", "baby shower pink cloud punch", "i did it my way tossed salad", "lubys cafeteria butternut brownie pie", "spiced poached pears", "lemon cajun stir fry", "iced banana cream", "potato ham onion chipotle soup", "chicken and penne casserole", "kahlua hot chocolate", "chicken and yoghurt curry", "oriental asparagus and mushrooms", "guacamole stuffed deviled eggs", "orzo with tomatoes feta and green onions", "kathy dessert baked bananas zwt ii asia", "hummus with pine nuts turkish style", "caramel delight", "whipped cream cream cheese frosting", "broccoli and cranberry salad", "raspberry lemonade", "pan broiled steak with whiskey sauce", "t g i fridays mudslide", "herb crusted fish fillets", "agua de valencia knock your socks off spanish cava punch", "orange brownie", "jiffy punch", "steak balmoral and whisky sauce from the witchery by the castle", "julies alabama white sauce", "ww potato gratin 5 points", "bo kaap cape malay curry powder south african spice mixture", "garlic herbed roasted red skin potatoes", "tasty broccoli salad", "risotto with pesto and mascarpone", "red potato and green bean saute", "caribbean sunset", "sriracha honey roasted broccoli", "salata marouli romaine lettuce salad", "goat cheese black olive mashed potatoes", "swirled cranberry cheesecake", "curried pea soup", "long island iced tea applebees tgi fridays style", "chocolate coconut dream bars", "bbq salmon filet", "blue margaritas", "sweet and sour pork balls", "spanish shrimp", "orange glazed pork chops", "heavenly lemon bread pudding", "spicy chicken soup with hints of lemongrass and coconut milk", "sweet onion and mashed potato bake", "smoky clam chowder", "cornish game hens with peach glaze", "garlic prime rib", "german apple cake with cream cheese frosting", "amazing sweet italian sausage pasta soup", "fresh orange slices with honey and cinnamon", "blackened tuna bites with cajun mustard", "tuna cobb salad", "greek shrimp with rigatoni", "creamy beet salad", "caponata eggplant and lots of good things", "lemon and oregano lamb loin chops", "pork chops with apples stuffing", "bacon spaghetti squash", "layered bean taco dip", "creamy lemon tarts", "strawberry and baileys fool", "italian style roast", "sourdough rosemary potato bread", "cracker barrel baby carrots", "portuguese tomato rice", "chocolate covered dipped strawberries", "caf a la russe chocolate coffee", "herbed potato with cottage cheese", "your basic tossed salad", "panzanella salad with bacon tomato and basil"},
	"snack":     {"hoisin marinated wing pieces", "feta marinated", "spicy roasted butternut seeds pumpkin seeds", "honey chipotle pecans", "best ever bruschetta", "body and soul health muffins", "kittencals beef burritos", "the traditional cyprus sandwich with halloumi onions and tomato", "delicious cheesy bacon and green onion potato skins", "fried almonds", "nifs peanut butter banana muffins", "lemon parsley popcorn", "not your ordinary chocolate chip cookies liqueur laced", "delicious and simple fruit dip", "fresh mango bread", "hot salty nuts", "omas boterkoek dutch buttercake", "apple butterflies", "lemon coconut muffins", "green mountain granola", "crispy fried chicken spring rolls", "guacamole stuffed deviled eggs", "hummus with a twist", "quick elephant ears", "homemade popsicles", "1 2 3 black bean salsa dip", "fantastic banana bran muffins", "sugared cinnamon almonds", "pumpkin chocolate chip muffins", "oatmeal cookie granola"},
	"dessert":   {"no bake hersheys bar pie", "big ol cowboy cookies", "crackle top molasses cookies", "old fashion oatmeal pie", "cranberry nut swirls", "butter balls", "peanut butter gooey cake", "mamas fruit cobbler", "pink stuff cherry pie filling pineapple dessert", "chocolate star cookies", "midsummer swedish strawberry compote  jordgubbskrm", "foolproof one bowl banana cake", "creamy apple dessert", "walnut chews", "yummy bread pudding", "white chocolate cheesecake", "hersheys kiss peanut butter cookies", "coconut pound cake", "frosted rhubarb cookies", "strawberry pavlova", "cookies n cream ice cream", "perfect pumpkin pie", "gluten free dutch sugar cookies", "raw apple crumble no bake", "cheesecake kugel", "moo less chocolate pie", "chocolate macadamia nut brownies", "disneyland snickerdoodles", "islands bananas foster", "frozen kahlua creme", "nifs peanut butter banana muffins", "peach cobbler with oatmeal cookie topping", "christmas cardamom butter cookies", "death by chocolate cake", "moms southern pecan pie", "the best brownies ever", "jerrys chocolate ice cream", "strawberry angel trifle", "zucchini mock apple pie", "low fat chocolate peanut butter dessert", "creamy raspberry mallow pie", "french napoleons", "pie crust cinnamon rolls", "not your ordinary chocolate chip cookies liqueur laced", "foolproof dark chocolate fudge", "whole wheat sugar cookies", "awesome kahlua cake", "up those antioxidants with blueberry sauce", "grammie millers swedish apple pie", "glendas flourless peanut butter cookies", "my best banana pudding dessert", "viskos praline sauce", "perfect purple punch", "reindeer bark", "lindas bloodshot eyeballs", "moroccan fruit salad", "apple dumpling bake", "simons pumpkin bread pudding", "baileys flourless peanut butter cookies", "a 1 cherry cobbler tart  a1", "monkey balls", "chocolate angel food cake", "creamy strawberries", "harvest cake", "deep dark chocolate moist cake", "spooktacular halloween graveyard cake", "cream cheese walnut drop cookies", "omas boterkoek dutch buttercake", "kates basic crepes", "banana spice bars", "ruth walls german apple cake", "low fat low cholesterol chocolate cake cupcakes", "lower fat peanut butter rice krispies bars", "nutella rolls", "fruit salad pudding", "strawberry and cream cheese parfait", "apple dessert quick", "betty crocker chocolate chip cookies 1971 mens favorites 22", "so there reeses peanut butter bars", "moms buttery apple cake", "chocolate almond roca bar", "turtles", "sesame toffee", "chocolate rainbow krispies treats", "dirt cups for kids", "ultimate seven layer bars", "raisin oat cookies", "snickers bar cookies", "french pie pastry", "sour cream pumpkin bundt cake", "microwave nut brittle", "cinnamon rolls buns", "nutella mousse", "blueberry sour cream cake", "angelic strawberry frozen yogurt", "chocolate chip bundt cake with chocolate glaze", "creole cake", "apricot banana squares", "banana snack cake with delicious cream cheese frosting", "pineapple coconut empanadas", "awesome chocolate butterscotch chip cookies", "easy homemade almond roca", "sonic strawberry cheesecake shake", "lubys cafeteria butternut brownie pie", "spiced poached pears", "chocolate mocha pudding  low carb", "iced banana cream", "kathy dessert  baked bananas zwt ii  asia", "whipped cream cream cheese frosting", "italian biscotti al la syd", "died and went to heaven chocolate cake diabetic version", "coffee and chocolate pudding", "mimis maine blueberry cobbler", "cherry cola float", "linzer bars", "confectioners sugar cookies", "double chocolate mint chip cookies", "quick elephant ears", "swirled cranberry cheesecake", "mexican rice pudding", "eclair torte", "spiced pumpkin pie", "caramel breakfast cake", "lime granita", "chocolate coconut dream bars", "blueberry banana pie", "grannys gingersnaps", "homemade popsicles", "heavenly lemon bread pudding", "pizzelles", "mckinley tea cakes", "lazy day cobbler", "old school deja vu chocolate peanut butter squares", "cheesecake pie", "aunt zanas amish sugar cookies eggless", "amish cream pie", "chocolate chip cookie dough ice cream", "snickerdoodles dream", "chocolate cheese fudge", "german apple cake with cream cheese frosting", "fresh orange slices with honey and cinnamon", "frozen oreo cookie dessert", "blueberry crunch", "amaretto bon bon balls", "red cherry pie", "creamy lemon tarts", "brownie truffles", "strawberry and baileys fool", "easy danish kringle", "chocolate covered dipped strawberries", "caf a la russe chocolate coffee"},
}
