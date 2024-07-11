Provide me with a JQ command that takes this JSON format/structure as input and:
 1. Marks duplicates by adding #1, #2, #3 and so on to their "name" field's value
 2. Adds a field called real_name after the "name" field that is the value of the filepath of download_url. (Convert the URL into a file path, forget about the protocol:// thingie and then use that. Remember to convert URI encoding to plain). This field should be NULL if the file path matches the "name" field

```
[
  {
    "name": "7z",
    "description": "Unarchiver",
    "download_url": "https://bin.ajam.dev/x86_64_Linux/7z",
    "size": "3.73 MB",
    "b3sum": "125acdc505ed6582ea1daec36c39d16749bbbf58ce2d19bdadaac27ff3b74f23",
    "sha256": "a2728a3dbd244cbb1a04f6a0998f53ec03abb7e3fb30e8e361fa22614c98e8d3",
    "build_date": "2024-06-24T03:39:33",
    "repo_url": "https://github.com/ip7z/7zip",
    "repo_author": "ip7z",
    "repo_info": "7-Zip",
    "repo_updated": "2024-07-05T14:39:53Z",
    "repo_released": "2024-06-19T10:45:51Z",
    "repo_version": "24.07",
    "repo_stars": "447",
    "repo_language": "C++",
    "repo_license": "",
    "repo_topics": "",
    "web_url": "https://www.7-zip.org",
    "extra_bins": ""
  },
```

If there were a duplicate of 7z, it would become:

```
[
  {
    "name": "7z#2",
    "description": "7z unarchiver but with additional support for PAX",
    "download_url": "https://bin.ajam.dev/x86_64_Linux/7z",
    "size": "3.73 MB",
    "b3sum": "125acdc505ed6582ea1daec36c39d16749bbbf58ce2d19bdadaac27ff3b74f23",
    "sha256": "a2728a3dbd244cbb1a04f6a0998f53ec03abb7e3fb30e8e361fa22614c98e8d3",
    "build_date": "2024-06-24T03:39:33",
    "repo_url": "https://github.com/ip7z/7zip",
    "repo_author": "ip7z",
    "repo_info": "7-Zip",
    "repo_updated": "2024-07-05T14:39:53Z",
    "repo_released": "2024-06-19T10:45:51Z",
    "repo_version": "24.07",
    "repo_stars": "447",
    "repo_language": "C++",
    "repo_license": "",
    "repo_topics": "",
    "web_url": "https://www.7-zip.org",
    "extra_bins": ""
  },
```

Since download_url points to a file which's file path is after removing the domain and the protocol is in the same dir as the value of the "name" field, the real_name is not added.
