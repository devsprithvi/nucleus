$version: "2"

namespace nucleus.sample.identity

structure User {
    @required
    id: UserId
    
    @required
    username: String
    
    @required
    email: String
    
    fullName: String
}

@pattern("^[a-zA-Z0-9-]+$")
string UserId
