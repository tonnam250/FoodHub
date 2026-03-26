-- Seed menus data (from menus (1).csv)
-- Assumes table: menus(id SERIAL PRIMARY KEY, name TEXT, price NUMERIC, image_url TEXT)

INSERT INTO menus (id, name, price, image_url) VALUES
  (1, 'Pad Thai', 95, 'https://media.istockphoto.com/id/1432057956/photo/pad-thai-noodles-with-chicken-on-dark-stone-table.jpg?s=612x612&w=0&k=20&c=BpYG-CpSo_nd7ZW2KhqNapb4HOK3qiWLNONevrhpoXk='),
  (2, 'Pad See Ew', 90, 'https://hot-thai-kitchen.com/wp-content/uploads/2023/04/pad-see-ew-sq-cu.jpg'),
  (3, 'Basil Chicken Rice', 70, 'https://theflavoursofkitchen.com/wp-content/uploads/2017/07/Thai-Basil-Chicken-3-scaled.jpg'),
  (4, 'Fried Rice', 75, 'https://cjeatsrecipes.com/wp-content/uploads/2023/11/Egg-Fried-Rice-in-a-bowl.jpg'),
  (5, 'Green Curry Chicken Rice', 85, 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQfY4n1cDL3UBnMrjIBM_puaCdYHjU7lYMXpQ&s'),
  (6, 'Massaman Curry Rice', 95, 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTWJGpK1xRNv3Mnx0nEmAx4KobNjqN77kKp5w&s'),
  (7, 'Tom Yum Noodle Soup', 80, 'https://www.seriouseats.com/thmb/74h5lyqDcKHGswmEVHFF8robIbs=/1500x0/filters:no_upscale():max_bytes(150000):strip_icc()/20231205-SEA-GuaydtiaaoTomYamMooSaap-AmandaSuarez-hero-450f7307a9dc4ca6b12fa06e606e8a01.jpg'),
  (8, 'Boat Noodle', 70, 'https://hot-thai-kitchen.com/wp-content/uploads/2014/09/boat-noodles-new-sq.jpg'),
  (10, 'Pork Rice Porridge', 60, 'https://rachelcooksthai.com/wp-content/uploads/2013/12/Jok-05.jpg'),
  (11, 'Chicken Rice', 75, 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQ_9XdAlHkr6USTTHzqaov7VdpLO4ofTphv6A&s'),
  (12, 'Crispy Pork Rice', 85, 'https://www.thammculture.com/wp-content/uploads/2022/02/Untitled-172.jpg'),
  (13, 'Omelette Rice', 60, 'https://f.ptcdn.info/689/035/000/1442885991-imagejpeg-o.jpg'),
  (14, 'Stir-fried Mixed Vegetables with Rice', 70, 'https://img.wongnai.com/p/1920x0/2019/02/05/37ff1888181e4cb1865f85b12896ea29.jpg'),
  (15, 'Grilled Pork with Sticky Rice', 65, 'https://www.thammculture.com/wp-content/uploads/2025/05/Untitled-639.jpg'),
  (16, 'Papaya Salad with Sticky Rice', 70, 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQVy_veuicqjUF2_ym8alon0j8iosY4MtRPmw&s'),
  (17, 'Thai Milk Tea', 40, 'https://p-u.popcdn.net/attachments/images/000/021/497/large/taiwan-milk-tea-with-bubble_1339-73160.jpg?1574675312'),
  (18, 'Lemon Tea', 35, 'https://grumpyrecipes.com/wp-content/uploads/2024/10/hong-kong-lemon-tea-recipe-feature.jpg'),
  (19, 'Iced Cocoa', 45, 'https://s359.kapook.com/pagebuilder/79c32820-a567-4cf0-a494-e297afdeabf5.jpg'),
  (20, 'Fresh Orange Juice', 50, 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcR2bIw4asa-RoIdb82YxQ8E_Xo_MCf5BPPh_Q&s')
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  price = EXCLUDED.price,
  image_url = EXCLUDED.image_url;

-- Keep sequence in sync with the max id
SELECT setval('menus_id_seq', (SELECT MAX(id) FROM menus));
