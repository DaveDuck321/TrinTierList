//Returns: {category:string, person1:{id:number, name:string, imgs:string[]}, ...}

let CurrentRank = {
    id: [],
    category: "",
}

function randomItem(list) {
    return list[Math.floor(Math.random()*list.length)];
}

async function ShowPeople(category) {
    document.getElementById("category").innerHTML = "Loading...";

    const result = await fetch(`?category=${category}`)
    const data = await result.json;

    document.getElementById("category").innerHTML = data.category;

    document.getElementById("person1-src").src = randomItem(data[0]);
    document.getElementById("person2-src").src = randomItem(data[1]);

    document.getElementById("person1-name").innerHTML = data.person1.name;
    document.getElementById("person2-name").innerHTML = data.person2.name;

    CurrentRank = {
        id: [data[0].id, data[1].id],
        category: result.category,
    };
}

async function Vote(id, category) {
    fetch("/vote", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json;charset=utf-8'
        },
        body: {
            type: 'vote',
            id: id,
            category: category,
        },
    });
    ShowPeople("random");
}

window.onload = ()=> {
    ShowPeople("random");

    document.getElementById("vote1").onclick = ()=>{Vote(CurrentRank.id[0], CurrentRank.category)};
    document.getElementById("vote2").onclick = ()=>{Vote(CurrentRank.id[1], CurrentRank.category)};
};