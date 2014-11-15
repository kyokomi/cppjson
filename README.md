cppjson
=======

A simple command-line tool for manipulating JSON for use in developing Cpp code.

## Usage

```
$ cat example/example.json | ./cppjson
```

```
$ curl -s https://raw.githubusercontent.com/kyokomi/cppjson/master/example/example.json | ./cppjson
```

## Input

```
{
    "MItem": [
        {
            "ItemID": 1,
            "ImageID": "1001",
            "Type": 0,
            "Name": "Testアイテム1",
            "Detail": "テスト用",
            "Param": 1
        }
    ],
    "MQuest": [
        {
            "QuestID": 1,
            "QuestName": "quest1",
            "QuestDetail": "quest1だよ",
            "FloorCount": 3
        },
        {
            "QuestID": 2,
            "QuestName": "quest2",
            "QuestDetail": "quest2だよ",
            "FloorCount": 5
        }
    ]
}
```

## Output

```cpp
struct Foo {
    
    struct MItem {
        std::string detail;
        std::string imageID;
        float itemID;
        std::string name;
        float param;
        float type;
    };
    std::vector<MItem> mItem;
    
    struct MQuest {
        float floorCount;
        std::string questDetail;
        float questID;
        std::string questName;
    };
    std::vector<MQuest> mQuest;
};
```

