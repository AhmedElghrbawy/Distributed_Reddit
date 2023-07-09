 CREATE TABLE IF NOT EXISTS Users (
 	handle VARCHAR(255) PRIMARY KEY,
 	display_name VARCHAR(255) NOT NULL,
 	avatar BYTEA NOT NULL,
 	karma INT,
 	created_at TIMESTAMP NOT NULL
 );
 
CREATE TABLE IF NOT EXISTS Subreddits (
	handle VARCHAR(255) PRIMARY KEY,
	title VARCHAR(255) NOT NULL,
	about VARCHAR(1000),
	avatar BYTEA NOT NULL,
	rules TEXT,
	created_at TIMESTAMP NOT NULL
);


CREATE TABLE IF NOT EXISTS User_Posts (
	id UUID PRIMARY KEY,
	title VARCHAR(255) NOT NULL UNIQUE,
	content VARCHAR(1000) NOT NULL,
	image BYTEA,
	created_at TIMESTAMP NOT NULL,
	number_of_votes INT,
	is_pinned BOOl,
	owner_handle VARCHAR(255) NOT NULL,
	subreddit_handle VARCHAR(255) NOT NULL,
	FOREIGN KEY (owner_handle) REFERENCES Users(handle)
); 

CREATE TABLE IF NOT EXISTS Subreddit_Posts (
 	id UUID PRIMARY KEY,
	title VARCHAR(255) NOT NULL UNIQUE,
	content VARCHAR(1000) NOT NULL,
	image BYTEA,
	created_at TIMESTAMP NOT NULL,
	number_of_votes INT,
	is_pinned BOOl,
	owner_handle VARCHAR(255) NOT NULL,
	subreddit_handle VARCHAR(255) NOT NULL,
	FOREIGN KEY (subreddit_handle) REFERENCES Subreddits(handle)
); 

CREATE TABLE IF NOT EXISTS User_Comments (
	id UUID PRIMARY KEY,
	content VARCHAR(1000) NOT NULL,
	image BYTEA,
	number_of_votes INT,
	owner_handle VARCHAR(255) NOT NULL,
	parent_comment_id UUID,
	post_id UUID NOT NULL,
	FOREIGN KEY (owner_handle) REFERENCES Users(handle),
	FOREIGN KEY (parent_comment_id) REFERENCES User_Comments(id),
	FOREIGN KEY (post_id) REFERENCES User_Posts(id)
);

CREATE TABLE IF NOT EXISTS Subreddit_Comments (
	id UUID PRIMARY KEY,
	content VARCHAR(1000) NOT NULL,
	image BYTEA,
	number_of_votes INT,
	owner_handle VARCHAR(255) NOT NULL,
	parent_comment_id UUID,
	post_id UUID NOT NULL,
	FOREIGN KEY (parent_comment_id) REFERENCES User_Comments(id),
	FOREIGN KEY (post_id) REFERENCES User_Posts(id)
);

CREATE TABLE IF NOT EXISTS Post_Tags (
	tag_name VARCHAR(255),
	post_id UUID,
	 
	PRIMARY KEY (tag_name, post_id)
);


CREATE TABLE IF NOT EXISTS User_Joined_Subreddits (
	user_handle VARCHAR(255),
	subreddit_handle VARCHAR(255),
	
	PRIMARY KEY (user_handle, subreddit_handle)
);


CREATE TABLE IF NOT EXISTS User_Admined_Subreddits (
	user_handle VARCHAR(255),
	subreddit_handle VARCHAR(255),
	
	PRIMARY KEY (user_handle, subreddit_handle)
);


 
CREATE TABLE IF NOT EXISTS Followage (
	follower_handle VARCHAR(255),
	followed_handle VARCHAR(255),
	followed_since TIMESTAMP,
	
	PRIMARY KEY (follower_handle, followed_handle)
);
