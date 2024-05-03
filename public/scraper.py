from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.common.keys import Keys
import time
import redis
import os
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Setup Redis connection
r = redis.Redis(
    host=os.getenv('REDIS_ADDR'),
    port=int(os.getenv('REDIS_PORT')),
    db=0,
    password=os.getenv('REDIS_PASSWORD')
)

url = "https://www.vanderbilt.edu/catalogs/kuali/undergraduate-23-24.php#/courses"
driver = webdriver.Chrome()
driver.get(url)
all_links = []

# Getting dropdown buttons
buttons = WebDriverWait(driver, 10).until(
    EC.presence_of_all_elements_located((By.CSS_SELECTOR, 'button.md-btn.md-btn--icon.md-pointer--hover.md-inline-block.style__collapseButton___12yNL'))
)

for button in buttons:
    button.click()
    time.sleep(0.3)

course_links = WebDriverWait(driver, 10).until(
    EC.presence_of_all_elements_located((By.CSS_SELECTOR, 'a[href^="#/courses/"]'))
)

all_links.extend(course_links)

for link in course_links:
        course_name = link.text.strip()
        course_link = link.get_attribute('href')
        # Open link in new tab
        driver.execute_script("window.open(arguments[0]);", course_link)
        driver.switch_to.window(driver.window_handles[1])

        try:
            course_description = WebDriverWait(driver, 10).until(
                EC.presence_of_element_located((By.CSS_SELECTOR, '.course-view__pre___2VF54 > div'))
            ).text
        except Exception as e:
            print(f"Failed to load course description for {course_name}: {str(e)}")
            course_description = "Description not available"

        # Split course name by first '-', then get rid of leading/trailing whitespace
        course_code, course_name = course_name.split('-', 1)
        course_code = course_code.strip()
        course_name = course_name.strip()
        print(course_name)

        # Store course details in Redis
        course_key = f"course:{course_code}"
        r.hset(course_key, mapping={
            'code': course_code,
            'name': course_name,
            'description': course_description,
            'link': course_link
        })

        # Going back to main tab
        driver.close()
        driver.switch_to.window(driver.window_handles[0])

driver.quit()
